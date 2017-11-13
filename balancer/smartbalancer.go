package balancer

import (
	"net"

	"time"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	log "github.com/sirupsen/logrus"
)

type Balancer struct {
	clients   map[string]net.Conn
	Scheduler *Scheduler

	listenCfg  string
	timeoutCfg time.Duration
	listener   net.Listener

	// Channels
	connect    chan (*core.Context)
	disconnect chan (net.Conn)
	stop       chan bool
}

func NewBalancer(listenCfg string) *Balancer {
	return &Balancer{
		Scheduler:  NewScheduler(),
		listenCfg:  listenCfg,
		timeoutCfg: 10 * time.Second,
		clients:    make(map[string]net.Conn),
		connect:    make(chan *core.Context),
		disconnect: make(chan net.Conn),
		stop:       make(chan bool),
	}
}

func (b *Balancer) Start() error {
	go func() {

		for {
			select {
			case ctx := <-b.connect:
				b.HandleClientConnect(ctx)

			case client := <-b.disconnect:
				b.HandleClientDisconnect(client)

			case <-b.stop:
				b.Stop()
				return
			}
		}
	}()

	// Scheduler takes care of selecting the right backend
	b.Scheduler.Start()

	if err := b.Listen(); err != nil {
		b.Stop()
		return err
	}

	return nil
}

func (b *Balancer) Stop() {
	b.Scheduler.Stop()

	log.Info("Shutting down load balancer. This will disconnect all clients")

	for _, conn := range b.clients {
		log.Debugf("Closing connection to client: %v", b.clients)
		conn.Close()
	}

	// Create new empty client list
	b.clients = make(map[string]net.Conn)
}

func (b *Balancer) HandleClientDisconnect(client net.Conn) {
	client.Close()
	delete(b.clients, client.RemoteAddr().String())
	b.Scheduler.stats.CurrentConnections <- uint(len(b.clients))
}

func (b *Balancer) HandleClientConnect(ctx *core.Context) {
	client := ctx.Conn

	b.clients[client.RemoteAddr().String()] = client
	b.Scheduler.stats.CurrentConnections <- uint(len(b.clients))

	go func() {
		b.handleConnection(ctx)
		b.disconnect <- client
	}()
}

func (b *Balancer) Listen() (err error) {
	b.listener, err = net.Listen("tcp", b.listenCfg)
	if err != nil {
		log.Error("Error starting listener on "+b.listenCfg, err)
		return err
	}

	go func() {
		for {
			conn, err := b.listener.Accept()
			if err != nil {
				log.Error(err)
				return
			}

			go b.wrap(conn)
		}
	}()

	log.Info("Started global listener on " + b.listenCfg)

	return nil
}

func (b *Balancer) wrap(conn net.Conn) {
	bufConn := core.NewBufferedConn(conn)
	hostname := core.HttpHostHeader(bufConn.Reader)
	log.Debugf("Hostname is: %v", hostname)

	b.connect <- &core.Context{
		Hostname: hostname,
		Conn:     bufConn,
	}
}



func (b *Balancer) handleConnection(ctx *core.Context) {
	clientConn := ctx.Conn

	log.Debug("Accepted connection from ", clientConn.RemoteAddr())

	// Find a router host that is healthy to forward the request to
	var err error
	routerHost, err := b.Scheduler.ElectRouterHostRequest(*ctx)
	if err != nil {
		log.Error(err, ". Closing connection: ", clientConn.RemoteAddr())
		return
	}

	log.Debugf("Selected target router host: %v", routerHost.HostIP)

	// Connect to router host
	routerHostConn, err := net.DialTimeout("tcp", routerHost.HostIP, b.timeoutCfg)
	bufferedRouterHostConn := core.NewBufferedConn(routerHostConn)
	if err != nil {
		b.Scheduler.IncrementRefused(routerHost.HostIP)
		log.Errorf("Error connecting to router host: %v. Err: %v", routerHost.HostIP, err)
		return
	}
	b.Scheduler.IncrementConnection(routerHost.HostIP)
	defer b.Scheduler.DecrementConnection(routerHost.HostIP)

	// Proxy the request & response bytes
	log.Debug("Begin ", clientConn.RemoteAddr(), " -> ", b.listener.Addr(), " -> ", bufferedRouterHostConn.RemoteAddr())
	cs := core.Proxy(clientConn, bufferedRouterHostConn, b.timeoutCfg)
	bs := core.Proxy(bufferedRouterHostConn, clientConn, b.timeoutCfg)

	isTx, isRx := true, true
	for isTx || isRx {
		select {
		case _, ok := <-cs:
			isRx = ok
		case _, ok := <-bs:
			isTx = ok
		}
	}

	log.Debug("End ", clientConn.RemoteAddr(), " -> ", b.listener.Addr(), " -> ", bufferedRouterHostConn.RemoteAddr())
}
