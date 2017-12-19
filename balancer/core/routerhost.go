package core

import (
	"strconv"
	"time"
)

type RouterHost struct {
	ClusterKey string
	HostIP     string
	HTTPPort   int
	HTTPSPort  int
	LastState  HostStats
	healthCheck *HealthCheck
}

func NewRouterHost(ip string, httpPort int, httpsPort int, s chan HealthCheckResult, clusterKey string) *RouterHost {
	rh := &RouterHost{
		ClusterKey: clusterKey,
		HostIP:     ip,
		HTTPPort:   httpPort,
		HTTPSPort:  httpsPort,
		LastState:  HostStats{},
	}

	rh.healthCheck = NewHealthCheck(rh, rh.HTTPPort, s, 1*time.Second)

	go rh.Start()

	return rh
}

func (rh *RouterHost) Key() string {
	return rh.HostIP + "-" + strconv.Itoa(rh.HTTPPort) + "/" + strconv.Itoa(rh.HTTPSPort)
}

func (rh *RouterHost) Start() {
	rh.healthCheck.Start()
}

func (rh *RouterHost) Stop() {
	rh.healthCheck.Stop()
}

