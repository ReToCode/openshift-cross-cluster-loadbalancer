package core

import (
	"time"
	"strconv"
)

type RouterHostStats struct {
	Healthy            bool   `json:"healthy"`
	TotalConnections   int64  `json:"totalConnections"`
	ActiveConnections  uint   `json:"activeConnections"`
	RefusedConnections uint64 `json:"refusedConnections"`
}

type RouterHost struct {
	ClusterKey string          `json:"clusterKey"`
	HostIP     string          `json:"hostIP"`
	HttpPort   int             `json:"httpPort"`
	HttpsPort  int             `json:"httpsPort"`
	Stats      RouterHostStats `json:"stats"`

	healthCheck *HealthCheck
}

func NewRouterHost(ip string, httpPort int, httpsPort int, s chan HealthCheckResult, clusterKey string) *RouterHost {
	rh := &RouterHost{
		ClusterKey: clusterKey,
		HostIP:     ip,
		HttpPort:   httpPort,
		HttpsPort:  httpsPort,
		Stats:      RouterHostStats{},
	}

	rh.healthCheck = NewHealthCheck(rh, rh.HttpPort, s, 1*time.Second)
	return rh
}

func (rh *RouterHost) Key() string {
	return rh.HostIP + "-" + strconv.Itoa(rh.HttpPort) + "/" + strconv.Itoa(rh.HttpsPort)
}

func (rh *RouterHost) Start() {
	rh.healthCheck.Start()
}

func (rh *RouterHost) Stop() {
	rh.healthCheck.Stop()
}
