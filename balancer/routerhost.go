package balancer

import "time"

type RouterHostStats struct {
	Healthy            bool   `json:"live"`
	TotalConnections   int64  `json:"total_connections"`
	ActiveConnections  uint   `json:"active_connections"`
	RefusedConnections uint64 `json:"refused_connections"`
}

type RouterHost struct {
	Stats RouterHostStats `json:"stats"`
	HostIP      string
	Routes      []string

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
