package balancer

import (
	"sync"

	"time"

	log "github.com/sirupsen/logrus"
)

type Cluster struct {
	name                string
	routerHosts         muxRouterHosts
	healthyHostCount    int
	healthChecksResults chan HealthCheckResult
}

type muxRouterHosts struct {
	mux sync.RWMutex
	m   map[string]Host
}

func NewCluster(name string) *Cluster {
	return &Cluster{
		name:                name,
		healthyHostCount:    0,
		healthChecksResults: make(chan HealthCheckResult),
		routerHosts:         muxRouterHosts{m: make(map[string]Host)},
	}
}

func (c *Cluster) Start() {
	// TODO:
	// Handle connections

	// Handle health check results
	go handleHealthCheckResults(c)
}

func (c *Cluster) AddRouterHost(ip string) {
	newHost := &RouterHost{
		hostIP:      ip,
		healthy:     false,
		healthCheck: NewHealthCheck(ip, c.healthChecksResults, 5*time.Second),
	}

	c.routerHosts.mux.Lock()
	c.routerHosts.m[ip] = newHost
	c.routerHosts.mux.Unlock()

	log.Infof("New router host was added: %v", ip)
	go newHost.Start()
}

func handleHealthCheckResults(c *Cluster) {
	for {
		select {
		case res := <-c.healthChecksResults:
			log.Debugf("Got new health check result of %v. %v", res.routerHostIP, res.healthy)

			c.routerHosts.mux.Lock()

			r := c.routerHosts.m[res.routerHostIP]

			// Healthy > not healthy
			if r.Healthy() && !res.healthy {
				c.healthyHostCount--
				log.Warningf("Router host %v degraded. Healthy host count: %v", res.routerHostIP, c.healthyHostCount)
			}

			// Not healthy > healthy
			if !r.Healthy() && res.healthy {
				c.healthyHostCount++
				log.Infof("Router host became healthy %v. Healthy host count: %v", res.routerHostIP, c.healthyHostCount)
			}

			// Update state
			r.SetHealth(res.healthy)

			c.routerHosts.mux.Unlock()
		}
	}
}
