package balancer

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type Stats struct {
	HealthyHostCount int

	Connections chan uint
}

type OperationAction int

const (
	IncrementConnection OperationAction = iota
	DecrementConnection
	IncrementRefused
	IncrementTx
	IncrementRx
)

type Operation struct {
	routerHostIp string
	action       OperationAction
}

type ElectRequest struct {
	Context  Context
	Response chan RouterHost
	Err      chan error
}

type Scheduler struct {
	routerHosts map[string]*RouterHost

	// Chanel for check results
	healthCheckResults chan HealthCheckResult

	balancer LeastConnectionsBalancer
	stats    Stats

	operations chan Operation
	elect      chan ElectRequest
	stop       chan bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		healthCheckResults: make(chan HealthCheckResult),
		routerHosts:        make(map[string]*RouterHost),
		balancer:           LeastConnectionsBalancer{},
		stats: Stats{
			HealthyHostCount: 0,
			Connections:      make(chan uint),
		},
		operations: make(chan Operation),
		elect:      make(chan ElectRequest),
		stop:       make(chan bool),
	}
}

func (s *Scheduler) Start() {
	log.Info("Starting scheduler")

	go func() {
		for {
			select {
			case connections := <-s.stats.Connections:
				log.Infof("Current connection count is: %v", connections)

			case checkResult := <-s.healthCheckResults:
				s.handleHealthCheckResults(checkResult)

			case op := <-s.operations:
				s.handleOperation(op)

			case electReq := <-s.elect:
				s.handleRouterHostElect(electReq)

			case <-s.stop:
				log.Info("Stopping scheduler")
				for _, rh := range s.routerHosts {
					rh.Stop()
				}
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.stop <- true
}

func (s *Scheduler) AddRouterHost(ip string) {
	newHost := &RouterHost{
		hostIP:      ip,
		healthCheck: NewHealthCheck(ip, s.healthCheckResults, 5*time.Second),
		Stats:       RouterHostStats{},
	}

	s.routerHosts[ip] = newHost

	log.Infof("New router host was added: %v", ip)

	// Start health checks of new host
	go newHost.Start()
}

func (s *Scheduler) IncrementRefused(routerHostIp string) {
	s.operations <- Operation{routerHostIp, IncrementRefused}
}

func (s *Scheduler) IncrementConnection(routerHostIp string) {
	s.operations <- Operation{routerHostIp, IncrementConnection}
}

func (s *Scheduler) DecrementConnection(routerHostIp string) {
	s.operations <- Operation{routerHostIp, DecrementConnection}
}

func (s *Scheduler) IncrementRx(routerHostIp string, c uint) {
	s.operations <- Operation{routerHostIp, IncrementRx}
}

func (s *Scheduler) IncrementTx(routerHostIp string, c uint) {
	s.operations <- Operation{routerHostIp, IncrementTx}
}

func (s *Scheduler) TakeRouterHost(ctx Context) (*RouterHost, error) {
	r := ElectRequest{ctx, make(chan RouterHost), make(chan error)}

	// Send election request
	s.elect <- r

	select {
	case err := <-r.Err:
		return nil, err

	case routerHost := <-r.Response:
		return &routerHost, nil
	}
}

func (s *Scheduler) handleOperation(op Operation) {
	switch op.action {
	case IncrementTx:
		//s.StatsHandler.Traffic <- core.ReadWriteCount{CountWrite: op.param.(uint), Target: op.target}
		return
	case IncrementRx:
		//this.StatsHandler.Traffic <- core.ReadWriteCount{CountRead: op.param.(uint), Target: op.target}
		return
	}

	routerHost, ok := s.routerHosts[op.routerHostIp]
	if !ok {
		log.Warn("Trying operation ", op.action, " on not tracket router host ip: ", op.routerHostIp)
		return
	}

	switch op.action {
	case IncrementRefused:
		routerHost.Stats.RefusedConnections++
	case IncrementConnection:
		routerHost.Stats.ActiveConnections++
		routerHost.Stats.TotalConnections++
	case DecrementConnection:
		routerHost.Stats.ActiveConnections--
	default:
		log.Warn("Don't know how to handle action ", op.action)
	}
}

func (s *Scheduler) handleRouterHostElect(req ElectRequest) {
	var routerHosts []*RouterHost
	for _, rh := range s.routerHosts {

		if !rh.Stats.Healthy {
			continue
		}

		routerHosts = append(routerHosts, rh)
	}

	// Elect RouterHost
	rh, err := s.balancer.GetRouterHost(req.Context, routerHosts)
	if err != nil {
		req.Err <- err
		return
	}

	req.Response <- *rh
}

func (s *Scheduler) handleHealthCheckResults(res HealthCheckResult) {
	r := s.routerHosts[res.routerHostIP]

	// Healthy > not healthy
	if r.Stats.Healthy && !res.healthy {
		s.stats.HealthyHostCount--
		log.Warningf("Router host %v degraded. Healthy host count: %v", res.routerHostIP, s.stats.HealthyHostCount)
	}

	// Not healthy > healthy
	if !r.Stats.Healthy && res.healthy {
		s.stats.HealthyHostCount++
		log.Infof("Router host became healthy %v. Healthy host count: %v", res.routerHostIP, s.stats.HealthyHostCount)
	}

	// Update state
	r.Stats.Healthy = res.healthy
}
