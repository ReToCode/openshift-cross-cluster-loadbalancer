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
	HostIP    string          `json:"hostIP"`
	HttpPort  int             `json:"httpPort"`
	HttpsPort int             `json:"httpsPort"`
	Stats     RouterHostStats `json:"stats"`
	Routes    []string        `json:"-"`

	healthCheck *HealthCheck
}



func NewRouterHost(ip string, httpPort int, httpsPort int, routes []string, s chan HealthCheckResult) *RouterHost {
	rh := &RouterHost{
		HostIP:    ip,
		HttpPort:  httpPort,
		HttpsPort: httpsPort,
		Stats:     RouterHostStats{},
		Routes:    routes,
	}

	rh.healthCheck =  NewHealthCheck(rh.Key(), rh.HostIP, rh.HttpPort, s, 1*time.Second)
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
