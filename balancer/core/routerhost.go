package core

import (
	"time"
)

type RouterHost struct {
	ClusterKey  string
	Name      string `json:"name"`
	HostIP      string `json:"hostIP"`
	HTTPPort    int    `json:"httpPort"`
	HTTPSPort   int    `json:"httpsPort"`
	LastState   HostStats
	healthCheck *HealthCheck
}

func NewRouterHost(name string, ip string, httpPort int, httpsPort int, s chan HealthCheckResult, clusterKey string) *RouterHost {
	rh := &RouterHost{
		Name: name,
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

func (rh *RouterHost) Start() {
	rh.healthCheck.Start()
}

func (rh *RouterHost) Stop() {
	rh.healthCheck.Stop()
}
