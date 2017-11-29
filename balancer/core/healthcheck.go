package core

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"strconv"
)

type HealthCheckResult struct {
	RouterHost *RouterHost
	Healthy    bool
}

type HealthCheck struct {
	routerHost *RouterHost
	checkPort  int

	interval time.Duration
	ticker   time.Ticker
	stop     chan bool
	status   chan HealthCheckResult
}

func NewHealthCheck(routerHost *RouterHost, checkPort int,
	status chan HealthCheckResult, checkInterval time.Duration) *HealthCheck {

	return &HealthCheck{
		routerHost: routerHost,
		checkPort:  checkPort,

		stop:     make(chan bool),
		status:   status,
		interval: checkInterval,
	}
}

func (hc *HealthCheck) Start() {
	logrus.Infof("Starting health checks for router host %v:%v", hc.routerHost.HostIP, hc.checkPort)

	hc.ticker = *time.NewTicker(hc.interval)

	go func() {
		for {
			select {
			case <-hc.ticker.C:
				go checkRouterHost(hc)

			case <-hc.stop:
				logrus.Debugf("Got stop signal for health check ticker.")
				hc.ticker.Stop()
				return
			}
		}
	}()

	select {}
}

func (hc *HealthCheck) Stop() {
	logrus.Infof("Stopping health checks for router host %v", hc.routerHost.HostIP)
	hc.stop <- true
}

func checkRouterHost(hc *HealthCheck) {
	conn, err := net.DialTimeout("tcp", hc.routerHost.HostIP+":"+strconv.Itoa(hc.checkPort), 5*time.Second)

	var healthy bool
	if err != nil {
		healthy = false
	} else {
		healthy = true
		conn.Close()
	}

	// Tell the balancer about the health result
	hc.status <- HealthCheckResult{
		RouterHost: hc.routerHost,
		Healthy:    healthy,
	}
}
