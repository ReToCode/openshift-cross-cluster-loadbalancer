package balancer

import (
	"net"

	"time"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/sirupsen/logrus"
	"strconv"
)

type BalancerConfig struct {
	httpListen        string
	httpsListen       string
	//proxyTimeout      time.Duration
	routerHostTimeout time.Duration
}

type Balancer struct {
	clients   map[string]net.Conn
	Scheduler *Scheduler

	cfg           BalancerConfig
	httpListener  net.Listener
	httpsListener net.Listener

	// Channels
	connect    chan *core.Context
	disconnect chan net.Conn
	stop       chan bool
}

func NewBalancer(httpListenCfg string, httpsListenCfg string) *Balancer {
	return &Balancer{
		Scheduler: NewScheduler(),
		cfg: BalancerConfig{
			httpListen:        httpListenCfg,
			httpsListen:       httpsListenCfg,
			//proxyTimeout:      10 * time.Second,
			routerHostTimeout: 5 * time.Second,
		},
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

	if err := b.ListenHttps(); err != nil {
		b.Stop()
		return err
	}
	if err := b.ListenHttp(); err != nil {
		b.Stop()
		return err
	}

	return nil
}

func (b *Balancer) Stop() {
	b.Scheduler.Stop()

	logrus.Info("Shutting down load balancer. This will disconnect all clients")

	for _, conn := range b.clients {
		logrus.Debugf("Closing connection to client: %v", b.clients)
		conn.Close()
	}

	// Create new empty client list
	b.clients = make(map[string]net.Conn)
}

func (b *Balancer) HandleClientDisconnect(client net.Conn) {
	client.Close()
	delete(b.clients, client.RemoteAddr().String())
	b.Scheduler.StatsHandler.Connections <- uint(len(b.clients))
}

func (b *Balancer) HandleClientConnect(ctx *core.Context) {
	client := ctx.Conn

	b.clients[client.RemoteAddr().String()] = client
	b.Scheduler.StatsHandler.Connections <- uint(len(b.clients))

	go func() {
		b.handleConnection(ctx)
		b.disconnect <- client
	}()
}

func (b *Balancer) ListenHttps() (err error) {
	b.httpsListener, err = net.Listen("tcp", b.cfg.httpsListen)
	if err != nil {
		logrus.Error("Error starting https listener on "+b.cfg.httpsListen, err)
		return err
	}

	go func() {
		for {
			conn, err := b.httpsListener.Accept()
			if err != nil {
				logrus.Error(err)
				return
			}

			go b.wrapHttpsConnection(conn)
		}
	}()

	logrus.Info("Started global https listener on " + b.cfg.httpsListen)

	return nil
}

func (b *Balancer) ListenHttp() (err error) {
	b.httpListener, err = net.Listen("tcp", b.cfg.httpListen)
	if err != nil {
		logrus.Error("Error starting http listener on "+b.cfg.httpListen, err)
		return err
	}

	go func() {
		for {
			conn, err := b.httpListener.Accept()
			if err != nil {
				logrus.Error(err)
				return
			}

			go b.wrapHttpConnection(conn)
		}
	}()

	logrus.Info("Started global http listener on " + b.cfg.httpListen)

	return nil
}

func (b *Balancer) wrapHttpsConnection(conn net.Conn) {
	// Get hostname based on SNI protocol
	sniConn, hostname, err := core.Sniff(conn, 5*time.Second)
	if err != nil {
		logrus.Error("Failed to get / parse ClientHello for sni: ", err)
		conn.Close()
		return
	}
	logrus.Debugf("Hostname is: %v", hostname)

	b.connect <- &core.Context{
		Hostname: hostname,
		HTTPS:    true,
		Conn:     core.NewBufferedConn(sniConn),
	}
}

func (b *Balancer) wrapHttpConnection(conn net.Conn) {
	// Get hostname out of http host header or take host value
	bufConn := core.NewBufferedConn(conn)
	hostname := core.HttpHostHeader(bufConn.Reader)
	logrus.Debugf("Hostname is: %v", hostname)

	b.connect <- &core.Context{
		Hostname: hostname,
		HTTPS:    false,
		Conn:     core.NewBufferedConn(bufConn),
	}
}

func (b *Balancer) handleConnection(ctx *core.Context) {
	clientConn := ctx.Conn

	logrus.Debug("Accepted connection from ", clientConn.RemoteAddr())

	// Find a router host that is healthy to forward the request to
	var err error
	routerHost, err := b.Scheduler.ElectRouterHostRequest(*ctx)
	if err != nil {
		logrus.Error(err, ". Closing connection: ", clientConn.RemoteAddr())
		return
	}

	var port int
	if ctx.HTTPS {
		port = routerHost.HTTPSPort
	} else {
		port = routerHost.HTTPPort
	}

	logrus.Debugf("Selected target router host: %v in port %v", routerHost.Name, port)

	// Connect to router host
	routerHostConn, err := net.DialTimeout("tcp", routerHost.HostIP+":"+strconv.Itoa(port), b.cfg.routerHostTimeout)
	bufferedRouterHostConn := core.NewBufferedConn(routerHostConn)
	if err != nil {
		b.Scheduler.UpdateRouterStats(routerHost.ClusterKey, routerHost.Name, IncrementRefused)
		logrus.Errorf("Error connecting to router host: %v. Err: %v", routerHost.Name, err)
		return
	}
	b.Scheduler.UpdateRouterStats(routerHost.ClusterKey, routerHost.Name, IncrementConnection)
	defer b.Scheduler.UpdateRouterStats(routerHost.ClusterKey, routerHost.Name, DecrementConnection)

	// Proxy the request & response bytes
	doneRxChan := core.Proxy(clientConn, bufferedRouterHostConn)
	doneTxChan := core.Proxy(bufferedRouterHostConn, clientConn)

	isTx, isRx := true, true
	for isTx || isRx {
		select {
		case <-doneRxChan:
			isRx = false
		case <-doneTxChan:
			isTx = false
		}
	}
}
