package balancer

import (
	"time"

	"sort"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/strategy"
	"github.com/sirupsen/logrus"
)

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

type RouterHostOperation struct {
	isAdd bool
	rh    *core.RouterHost
}

type ElectRequest struct {
	Context  core.Context
	Response chan core.RouterHost
	Err      chan error
}

type Scheduler struct {
	routerHosts        map[string]*core.RouterHost
	balancer           strategy.LeastConnectionsBalancer
	currentConnections uint

	StatsUpdate        chan core.GlobalStats
	healthCheckResults chan core.HealthCheckResult
	hostOperation      chan RouterHostOperation
	statsOperation     chan StatsOperation
	elect              chan ElectRequest
	connections        chan uint
	stop               chan bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		routerHosts:        make(map[string]*core.RouterHost),
		balancer:           strategy.LeastConnectionsBalancer{},
		currentConnections: 0,

		StatsUpdate:        make(chan core.GlobalStats),
		healthCheckResults: make(chan core.HealthCheckResult),
		hostOperation:      make(chan RouterHostOperation),
		statsOperation:     make(chan StatsOperation),
		elect:              make(chan ElectRequest),
		connections:        make(chan uint),
		stop:               make(chan bool),
	}
}

func (s *Scheduler) Start() {
	statsPushTicker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case conn := <-s.connections:
				s.currentConnections = conn

			case <-statsPushTicker.C:
				s.updateStats()

			case op := <-s.hostOperation:
				s.handleRouterHostOperation(op)

			case checkResult := <-s.healthCheckResults:
				s.handleHealthCheckResults(checkResult)

			case op := <-s.statsOperation:
				s.handleRouterHostStats(op)

			case electReq := <-s.elect:
				s.handleRouterHostElect(electReq)

			case <-s.stop:
				logrus.Info("Stopping scheduler")

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
	newHost := core.NewRouterHost(ip, routes, s.healthCheckResults)

	s.hostOperation <- RouterHostOperation{
		true, newHost,
	}
}

func (s *Scheduler) IncrementRefused(routerHostIp string) {
	s.statsOperation <- StatsOperation{routerHostIp, IncrementRefused}
}

func (s *Scheduler) IncrementConnection(routerHostIp string) {
	s.statsOperation <- StatsOperation{routerHostIp, IncrementConnection}
}

func (s *Scheduler) DecrementConnection(routerHostIp string) {
	s.statsOperation <- StatsOperation{routerHostIp, DecrementConnection}
}

func (s *Scheduler) ElectRouterHostRequest(ctx core.Context) (*core.RouterHost, error) {
	r := ElectRequest{ctx, make(chan core.RouterHost), make(chan error)}

	// Send election request
	s.elect <- r

	select {
	case err := <-r.Err:
		return nil, err

	case routerHost := <-r.Response:
		return &routerHost, nil
	}
}

func (s *Scheduler) handleRouterHostOperation(op RouterHostOperation) {
	if op.isAdd {
		s.routerHosts[op.rh.HostIP] = op.rh

		// Start health checks of router host
		go s.routerHosts[op.rh.HostIP].Start()

		logrus.Infof("New router host was added: %v to scheduler", op.rh.HostIP)
	} else {
		logrus.Errorf("Deletion of hosts is not yet possible")
	}

	s.updateStats()
}

func (s *Scheduler) updateStats() {
	// Create a sorted list for the UI
	hostList := []core.RouterHost{}
	for _, rh := range s.routerHosts {
		hostList = append(hostList, *rh)
	}

	sort.Slice(hostList, func(i, j int) bool {
		return hostList[i].HostIP < hostList[j].HostIP
	})

	// Tell the UI about it
	s.StatsUpdate <- core.GlobalStats{
		Mutation: "uiStats",
		HostList: hostList,
		CurrentConnections: s.currentConnections,
	}
}

func (s *Scheduler) handleRouterHostStats(op StatsOperation) {
	routerHost, ok := s.routerHosts[op.routerHostIp]
	if !ok {
		logrus.Warn("Trying operation ", op.action, " on not tracked router host ip: ", op.routerHostIp)
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
		logrus.Warn("Don't know how to handle action ", op.action)
	}
}

func (s *Scheduler) handleRouterHostElect(req ElectRequest) {
	var hosts []*core.RouterHost
	var healthyHosts []*core.RouterHost

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
		logrus.Warnf("Route %v has no valid target. Balancing to all healthy router hosts", req.Context.Hostname)
		hosts = healthyHosts
	}

	if req.Context.Hostname == "" {
		logrus.Warnf("No route name was parsed. Balancing to all healthy router hosts")
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

func (s *Scheduler) handleHealthCheckResults(res core.HealthCheckResult) {
	r := s.routerHosts[res.RouterHostIP]

	// Healthy > not healthy
	if r.Stats.Healthy && !res.Healthy {
		logrus.Warningf("Router host %v degraded", res.RouterHostIP)
	}

	// Not healthy > healthy
	if !r.Stats.Healthy && res.Healthy {
		logrus.Infof("Router host became healthy %v", res.RouterHostIP)
	}

	// Update state
	r.Stats.Healthy = res.Healthy
}
