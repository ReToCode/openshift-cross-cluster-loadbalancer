package balancer

type Host interface {
	Start()
	Stop()
	Healthy() bool
	SetHealth(health bool)
}

type RouterHost struct {
	hostIP      string
	healthy     bool
	healthCheck *HealthCheck
}

func (rh *RouterHost) Start() {
	// TODO: Handle connections

	rh.healthCheck.Start()
}

func (rh *RouterHost) Stop() {
	// TODO: Undo start

	rh.healthCheck.Stop()
}

func (rh *RouterHost) Healthy() bool {
	return rh.healthy
}

func (rh *RouterHost) SetHealth(health bool) {
	rh.healthy = health
}