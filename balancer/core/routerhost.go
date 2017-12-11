package core

import (
	"strconv"
	"time"
)

type RouterHostStats struct {
	Healthy            bool   `json:"healthy"`
	TotalConnections   int64  `json:"totalConnections"`
	ActiveConnections  uint   `json:"activeConnections"`
	RefusedConnections uint64 `json:"refusedConnections"`
}

type RouterHost struct {
	ClusterKey string            `json:"clusterKey"`
	HostIP     string            `json:"hostIP"`
	HttpPort   int               `json:"httpPort"`
	HttpsPort  int               `json:"httpsPort"`
	LastState  RouterHostStats   `json:"-"`
	Stats      []RouterHostStats `json:"stats"`

	healthCheck *HealthCheck
}

func NewRouterHost(ip string, httpPort int, httpsPort int, s chan HealthCheckResult, clusterKey string) *RouterHost {
	rh := &RouterHost{
		ClusterKey: clusterKey,
		HostIP:     ip,
		HttpPort:   httpPort,
		HttpsPort:  httpsPort,
		LastState:  RouterHostStats{},
		Stats:      []RouterHostStats{},
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

func (rh *RouterHost) ResetStats() {
	rh.LastState.TotalConnections = 0
}

func (rh *RouterHost) UpdateStats() {
	if len(rh.Stats) >= MaxTicks {
		rh.Stats = rh.Stats[1:]
	} else {
		for i :=0; i <= MaxTicks; i++ {
			rh.Stats = append(rh.Stats, RouterHostStats{})
		}
	}
	rh.Stats = append(rh.Stats, RouterHostStats{
		TotalConnections: rh.LastState.TotalConnections,
		Healthy: rh.LastState.Healthy,
		ActiveConnections: rh.LastState.ActiveConnections,
		RefusedConnections: rh.LastState.RefusedConnections,
	})

	// Reset RefusedConnections since last tick
	rh.LastState.RefusedConnections = 0
}
