package balancer

import (
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	log "github.com/sirupsen/logrus"
)

type Stats struct {
	HealthyHostCount int

	CurrentConnections chan uint
}

type StatsOperationAction int

const (
	IncrementConnection StatsOperationAction = iota
	DecrementConnection
	IncrementRefused
)

type StatsOperation struct {
	routerHostIp string
	action       StatsOperationAction
}

type ElectRequest struct {
	Context  core.Context
	Response chan RouterHost
	Err      chan error
}

type Scheduler struct {
	routerHosts map[string]*RouterHost

	healthCheckResults chan HealthCheckResult

	balancer LeastConnectionsBalancer
	stats    Stats

	operations chan StatsOperation
	elect      chan ElectRequest
	stop       chan bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		healthCheckResults: make(chan HealthCheckResult),
		routerHosts:        make(map[string]*RouterHost),
		balancer:           LeastConnectionsBalancer{},
		stats: Stats{
			HealthyHostCount:   0,
			CurrentConnections: make(chan uint),
		},
		operations: make(chan StatsOperation),
		elect:      make(chan ElectRequest),
		stop:       make(chan bool),
	}
}

func (s *Scheduler) Start() {
	log.Info("Starting scheduler")

	go func() {
		for {
			select {
			case connections := <-s.stats.CurrentConnections:
				log.Infof("Current connections: %v", connections)

			case checkResult := <-s.healthCheckResults:
				s.handleHealthCheckResults(checkResult)

			case op := <-s.operations:
				s.updateStats(op)

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

func (s *Scheduler) AddRouterHost(ip string, routes []string) {
	newHost := NewRouterHost(ip, routes, s.healthCheckResults)

	s.routerHosts[ip] = newHost

	// Start health checks of router host
	go newHost.Start()

	log.Infof("New router host was added: %v to scheduler", ip)
}

func (s *Scheduler) IncrementRefused(routerHostIp string) {
	s.operations <- StatsOperation{routerHostIp, IncrementRefused}
}

func (s *Scheduler) IncrementConnection(routerHostIp string) {
	s.operations <- StatsOperation{routerHostIp, IncrementConnection}
}

func (s *Scheduler) DecrementConnection(routerHostIp string) {
	s.operations <- StatsOperation{routerHostIp, DecrementConnection}
}

func (s *Scheduler) ElectRouterHostRequest(ctx core.Context) (*RouterHost, error) {
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

func (s *Scheduler) updateStats(op StatsOperation) {
	routerHost, ok := s.routerHosts[op.routerHostIp]
	if !ok {
		log.Warn("Trying operation ", op.action, " on not tracked router host ip: ", op.routerHostIp)
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
	var hosts []*RouterHost
	var healthyHosts []*RouterHost

	for _, rh := range s.routerHosts {
		// 1. Check if healthy
		if !rh.Stats.Healthy {
			continue
		}

		healthyHosts = append(healthyHosts, rh)

		// 2. Check if route is handled by current router host
		if req.Context.Hostname != "" {
			for _, r := range rh.Routes {
				if r == req.Context.Hostname {
					hosts = append(hosts, rh)
				}
			}
		}
	}

	if req.Context.Hostname != "" && len(hosts) == 0 {
		log.Warnf("Route %v has no valid target. Balancing to all healthy router hosts", req.Context.Hostname)
		hosts = healthyHosts
	}

	if req.Context.Hostname == "" {
		log.Warnf("No route name was parsed. Balancing to all healthy router hosts")
		hosts = healthyHosts
	}

	// Elect RouterHost
	rh, err := s.balancer.GetRouterHost(req.Context, hosts)
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
