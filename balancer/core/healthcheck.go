package core

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type HealthCheckResult struct {
	RouterHostIP string
	Healthy      bool
}

type HealthCheck struct {
	interval     time.Duration
	ticker       time.Ticker
	stop         chan bool
	status       chan HealthCheckResult
	routerHostIP string
}

func NewHealthCheck(ip string, status chan HealthCheckResult, checkInterval time.Duration) *HealthCheck {
	return &HealthCheck{
		routerHostIP: ip,
		stop:         make(chan bool),
		status:       status,
		interval:     checkInterval,
	}
}

func (hc *HealthCheck) Start() {
	logrus.Infof("Starting health checks for router host %v", hc.routerHostIP)

	hc.ticker = *time.NewTicker(hc.interval)

	go func() {
		for {
			select {
			case <-hc.ticker.C:
				go checkRouterHost(hc.routerHostIP, hc.status)

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
	logrus.Infof("Stopping health checks for router host %v", hc.routerHostIP)
	hc.stop <- true
}

func checkRouterHost(routerHostIp string, result chan<- HealthCheckResult) {
	conn, err := net.DialTimeout("tcp", routerHostIp, 5*time.Second)

	var healthy bool
	if err != nil {
		healthy = false
	} else {
		healthy = true
		conn.Close()
	}

	// Tell the balancer about the health result
	result <- HealthCheckResult{
		RouterHostIP: routerHostIp,
		Healthy:      healthy,
	}
}
