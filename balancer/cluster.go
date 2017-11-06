package balancer

import (
	"net"

	log "github.com/sirupsen/logrus"
	"time"
)

type Cluster struct {
	name string

	clients   map[string]net.Conn
	Scheduler *Scheduler

	listener net.Listener

	// Channels
	connect    chan (*Context)
	disconnect chan (net.Conn)
	stop       chan bool
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		name:       name,
		Scheduler:  NewScheduler(),
		clients:    make(map[string]net.Conn),
		connect:    make(chan *Context),
		disconnect: make(chan net.Conn),
		stop:       make(chan bool),
	}
}

func (c *Cluster) Start() error {
	go func() {

		for {
			select {
			case client := <-c.disconnect:
				c.HandleClientDisconnect(client)

			case ctx := <-c.connect:
				c.HandleClientConnect(ctx)

			case <-c.stop:
				c.Stop()
				return
			}
		}
	}()

	c.Scheduler.Start()

	if err := c.Listen(); err != nil {
		c.Stop()
		return err
	}

	return nil
}

func (c *Cluster) Stop() {
	c.Scheduler.Stop()

	// Todo draining
	// Disconnect existing clients
	for _, conn := range c.clients {
		log.Debugf("Closing connection to client: %v", c.clients)
		conn.Close()
	}

	// Create new empty client list
	c.clients = make(map[string]net.Conn)
}

func (c *Cluster) HandleClientDisconnect(client net.Conn) {
	client.Close()
	delete(c.clients, client.RemoteAddr().String())
	c.Scheduler.stats.Connections <- uint(len(c.clients))
}

func (c *Cluster) HandleClientConnect(ctx *Context) {
	client := ctx.Conn

	c.clients[client.RemoteAddr().String()] = client
	c.Scheduler.stats.Connections <- uint(len(c.clients))

	go func() {
		c.handleConnection(ctx)
		c.disconnect <- client
	}()
}

func (c *Cluster) Listen() (err error) {
	c.listener, err = net.Listen("tcp", "localhost:9999")
	if err != nil {
		log.Error("Error starting listener on localhost:9999", err)
		return err
	}

	go func() {
		for {
			conn, err := c.listener.Accept()
			if err != nil {
				log.Error(err)
				return
			}

			go c.wrap(conn)
		}
	}()

	log.Info("Started global listener on localhost:9999")

	return nil
}

func (c *Cluster) wrap(conn net.Conn) {
	// Todo get hostname out of sni
	c.connect <- &Context{
		Hostname: "bla",
		Conn:     conn,
	}
}

func (c *Cluster) handleConnection(ctx *Context) {
	clientConn := ctx.Conn

	log.Debug("Accepted ", clientConn.RemoteAddr(), " -> ", c.name)

	// Find a router host that is healthy to forward the request to
	var err error
	routerHost, err := c.Scheduler.TakeRouterHost(*ctx)
	if err != nil {
		log.Error(err, "; Closing connection: ", clientConn.RemoteAddr())
		return
	}

	log.Debugf("Selected target router host: %v", routerHost.hostIP)

	// Connect to router host
	routerHostConn, err := net.DialTimeout("tcp", routerHost.hostIP, 10 * time.Second)
	if err != nil {
		c.Scheduler.IncrementRefused(routerHost.hostIP)
		log.Errorf("Error connecting to router host: %v. Err: %v", routerHost.hostIP, err)
		return
	}
	c.Scheduler.IncrementConnection(routerHost.hostIP)
	defer c.Scheduler.DecrementConnection(routerHost.hostIP)

	// Proxy the request & response bytes for tx stats
	log.Debug("Begin ", clientConn.RemoteAddr(), " -> ", c.listener.Addr(), " -> ", routerHostConn.RemoteAddr())
	cs := proxy(clientConn, routerHostConn, 10 * time.Second)
	bs := proxy(routerHostConn, clientConn, 10 * time.Second)

	isTx, isRx := true, true
	for isTx || isRx {
		select {
		case s, ok := <-cs:
			isRx = ok
			c.Scheduler.IncrementRx(routerHost.hostIP, s.CountWrite)
		case s, ok := <-bs:
			isTx = ok
			c.Scheduler.IncrementTx(routerHost.hostIP, s.CountWrite)
		}
	}

	log.Debug("End ", clientConn.RemoteAddr(), " -> ", c.listener.Addr(), " -> ", routerHostConn.RemoteAddr())
}
