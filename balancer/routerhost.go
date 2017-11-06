package balancer

type RouterHostStats struct {
	Healthy            bool   `json:"live"`
	TotalConnections   int64  `json:"total_connections"`
	ActiveConnections  uint   `json:"active_connections"`
	RefusedConnections uint64 `json:"refused_connections"`
	RxBytes            uint64 `json:"rx"`
	TxBytes            uint64 `json:"tx"`
	RxSecond           uint   `json:"rx_second"`
	TxSecond           uint   `json:"tx_second"`
}

type RouterHost struct {
	Stats       RouterHostStats `json:"stats"`
	hostIP      string
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
