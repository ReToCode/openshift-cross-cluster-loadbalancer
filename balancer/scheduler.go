package balancer

import (
	"time"

	"sync"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/balancing"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/stats"
	"github.com/sirupsen/logrus"
)

type StatsOperationAction int

const (
	IncrementConnection StatsOperationAction = iota
	DecrementConnection
	IncrementRefused
)

type ElectRequest struct {
	Context  core.Context
	Response chan core.RouterHost
	Err      chan error
}

type SafeClusters struct {
	v   map[string]*core.Cluster
	mux sync.Mutex
}

// Scheduler handles:
// - The overall state (clusters, router hosts)
// - Health-Check-Results
// - Election of target router hosts
type Scheduler struct {
	clusters     SafeClusters
	StatsHandler *stats.StatsHandler

	healthCheckResults chan core.HealthCheckResult
	elect              chan ElectRequest
	ResetStats         chan bool
	stop               chan bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		clusters:     SafeClusters{v: map[string]*core.Cluster{}},
		StatsHandler: stats.NewHandler(),

		healthCheckResults: make(chan core.HealthCheckResult),
		elect:              make(chan ElectRequest),
		ResetStats:         make(chan bool),
		stop:               make(chan bool),
	}
}

func (s *Scheduler) Start() {
	s.StatsHandler.Start()

	hostsPushTicket := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-s.ResetStats:
				s.resetStats()

			case checkResult := <-s.healthCheckResults:
				s.handleHealthCheckResults(checkResult)

			case electReq := <-s.elect:
				s.handleRouterHostElect(electReq)

			case <-hostsPushTicket.C:
				s.StatsHandler.RouterHosts <- s.routerHosts()
				s.resetRefusedStats()

			case <-s.stop:
				logrus.Info("Stopping scheduler")
				s.StatsHandler.Stop()

				s.clusters.mux.Lock()
				for _, c := range s.clusters.v {
					c.Stop()
				}
				s.clusters.mux.Unlock()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.stop <- true
}

func (s *Scheduler) AddCluster(key string, routes []core.Route) {
	logrus.Infof("Added cluster: %v", key)
	cluster := core.NewCluster(key, routes)

	s.clusters.mux.Lock()
	s.clusters.v[key] = cluster
	s.clusters.mux.Unlock()
}

func (s *Scheduler) AddRouterHost(clusterKey string, ip string, httpPort int, httpsPort int) {
	newHost := core.NewRouterHost(ip, httpPort, httpsPort, s.healthCheckResults, clusterKey)
	logrus.Infof("New router host was added: %v to scheduler", newHost.Key())

	s.clusters.mux.Lock()
	s.clusters.v[clusterKey].RouterHosts[newHost.Key()] = newHost
	s.clusters.mux.Unlock()
}

func (s *Scheduler) UpdateRouterStats(clusterKey string, routerHostKey string, action StatsOperationAction) {
	s.clusters.mux.Lock()
	defer s.clusters.mux.Unlock()

	routerHost, ok := s.clusters.v[clusterKey].RouterHosts[routerHostKey]
	if !ok {
		logrus.Warn("Trying operation ", action, " on not tracked router host ip: ", routerHostKey)
		return
	}

	switch action {
	case IncrementRefused:
		routerHost.LastState.RefusedConnections++
	case IncrementConnection:
		routerHost.LastState.ActiveConnections++
		routerHost.LastState.TotalConnections++
	case DecrementConnection:
		routerHost.LastState.ActiveConnections--
	default:
		logrus.Warn("Don't know how to handle action ", action)
	}
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

func (s *Scheduler) routerHosts() []core.RouterHost {
	s.clusters.mux.Lock()
	defer s.clusters.mux.Unlock()
	l := make([]core.RouterHost, 0)
	for _, c := range s.clusters.v {
		for _, rh := range c.RouterHosts {
			l = append(l, *rh)
		}
	}
	return l
}

func (s *Scheduler) resetStats() {
	s.clusters.mux.Lock()
	for _, cl := range s.clusters.v {
		for _, rh := range cl.RouterHosts {
			rh.LastState.TotalConnections = 0
		}
	}
	s.clusters.mux.Unlock()
}

func (s *Scheduler) resetRefusedStats(){
	s.clusters.mux.Lock()
	for _, cl := range s.clusters.v {
		for _, rh := range cl.RouterHosts {
			rh.LastState.RefusedConnections = 0
		}
	}
	s.clusters.mux.Unlock()
}

func (s *Scheduler) handleRouterHostElect(req ElectRequest) {
	s.clusters.mux.Lock()
	defer s.clusters.mux.Unlock()
	rh, err := balancing.ElectRouterHost(req.Context, s.clusters.v)
	if err != nil {
		req.Err <- err
		return
	}

	req.Response <- *rh
}

func (s *Scheduler) handleHealthCheckResults(res core.HealthCheckResult) {
	s.clusters.mux.Lock()
	defer s.clusters.mux.Unlock()

	// Healthy > not healthy
	if res.RouterHost.LastState.Healthy && !res.Healthy {
		logrus.Warningf("Router host %v on %v degraded", res.RouterHost.Key(), res.RouterHost.ClusterKey)
	}

	// Not healthy > healthy
	if !res.RouterHost.LastState.Healthy && res.Healthy {
		logrus.Infof("Router host %v on %v became healthy", res.RouterHost.Key(), res.RouterHost.ClusterKey)
	}

	// Update state
	res.RouterHost.LastState.Healthy = res.Healthy
}
