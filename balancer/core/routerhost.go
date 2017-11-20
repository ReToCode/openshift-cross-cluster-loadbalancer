package core

import "time"

type RouterHostStats struct {
	Healthy            bool   `json:"healthy"`
	TotalConnections   int64  `json:"totalConnections"`
	ActiveConnections  uint   `json:"activeConnections"`
	RefusedConnections uint64 `json:"refusedConnections"`
}

type RouterHost struct {
	Stats  RouterHostStats `json:"stats"`
	HostIP string          `json:"hostIP"`
	Routes []string        `json:"-"`

	healthCheck *HealthCheck
}

func NewRouterHost(ip string, routes []string, s chan HealthCheckResult) *RouterHost {
	return &RouterHost{
		HostIP:      ip,
		Routes:      routes,
		Stats:       RouterHostStats{},
		healthCheck: NewHealthCheck(ip, s, 1*time.Second),
	}
}

func (rh *RouterHost) Start() {
	rh.healthCheck.Start()
}

func (rh *RouterHost) Stop() {
	rh.healthCheck.Stop()
}
