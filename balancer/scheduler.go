package balancer

import (
	"time"

	"sort"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/balancing"
	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/sirupsen/logrus"
)

type StatsOperationAction int

const (
	IncrementConnection StatsOperationAction = iota
	DecrementConnection
	IncrementRefused
)

type StatsOperation struct {
	clusterKey    string
	routerHostKey string
	action        StatsOperationAction
}

type RouterHostOperation struct {
	isAdd      bool
	routerHost *core.RouterHost
}

type ClusterOperation struct {
	isAdd   bool
	cluster *core.Cluster
}

type ElectRequest struct {
	Context  core.Context
	Response chan core.RouterHost
	Err      chan error
}

// Scheduler handles:
// - the overall state (clusters, router hosts)
// - Health-Check-Results
// - Election of target router hosts
type Scheduler struct {
	clusters           map[string]*core.Cluster
	currentConnections uint

	StatsUpdate        chan core.GlobalStats
	healthCheckResults chan core.HealthCheckResult
	hostOperation      chan RouterHostOperation
	clusterOperation   chan ClusterOperation
	statsOperation     chan StatsOperation
	elect              chan ElectRequest
	connections        chan uint
	stop               chan bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		clusters:           map[string]*core.Cluster{},
		currentConnections: 0,

		StatsUpdate:        make(chan core.GlobalStats),
		healthCheckResults: make(chan core.HealthCheckResult),
		hostOperation:      make(chan RouterHostOperation),
		clusterOperation:   make(chan ClusterOperation),
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

			case op := <-s.clusterOperation:
				s.handleClusterOperation(op)

			case checkResult := <-s.healthCheckResults:
				s.handleHealthCheckResults(checkResult)

			case op := <-s.statsOperation:
				s.handleRouterHostStats(op)

			case electReq := <-s.elect:
				s.handleRouterHostElect(electReq)

			case <-s.stop:
				logrus.Info("Stopping scheduler")

				for _, c := range s.clusters {
					c.Stop()
				}

				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.stop <- true
}

func (s *Scheduler) AddCluster(key string, routes []core.Route) {
	cluster := core.NewCluster(key, routes)

	s.clusterOperation <- ClusterOperation{
		true, cluster,
	}
}

func (s *Scheduler) AddRouterHost(clusterKey string, ip string, httpPort int, httpsPort int) {
	newHost := core.NewRouterHost(ip, httpPort, httpsPort, s.healthCheckResults, clusterKey)

	s.hostOperation <- RouterHostOperation{
		true, newHost,
	}
}

func (s *Scheduler) UpdateRouterStats(clusterKey string, routerHostKey string, action StatsOperationAction) {
	s.statsOperation <- StatsOperation{clusterKey, routerHostKey, action}
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
		s.clusters[op.routerHost.ClusterKey].RouterHosts[op.routerHost.Key()] = op.routerHost

		// Start health checks of router host
		go s.clusters[op.routerHost.ClusterKey].RouterHosts[op.routerHost.Key()].Start()

		logrus.Infof("New router host was added: %v to scheduler", op.routerHost.Key())
	} else {
		logrus.Errorf("Deletion of hosts is not yet possible")
	}

	s.updateStats()
}

func (s *Scheduler) handleClusterOperation(op ClusterOperation) {
	if op.isAdd {
		logrus.Infof("Added cluster: %v", op.cluster.Key)

		s.clusters[op.cluster.Key] = op.cluster
	} else {
		logrus.Errorf("Deletion of clusters is not yet possible")
	}
}

func (s *Scheduler) updateStats() {
	// Create a sorted list for the UI
	hostList := []core.RouterHost{}
	for _, cl := range s.clusters {
		for _, rh := range cl.RouterHosts {
			hostList = append(hostList, *rh)
		}
	}

	sort.Slice(hostList, func(i, j int) bool {
		return hostList[i].Key() < hostList[j].Key()
	})

	// Tell the UI about it
	s.StatsUpdate <- core.GlobalStats{
		Mutation:           "uiStats",
		HostList:           hostList,
		CurrentConnections: s.currentConnections,
	}
}

func (s *Scheduler) handleRouterHostStats(op StatsOperation) {
	routerHost, ok := s.clusters[op.clusterKey].RouterHosts[op.routerHostKey]
	if !ok {
		logrus.Warn("Trying operation ", op.action, " on not tracked router host ip: ", op.routerHostKey)
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
	rh, err := balancing.ElectRouterHost(req.Context, s.clusters)
	if err != nil {
		req.Err <- err
		return
	}

	req.Response <- *rh
}

func (s *Scheduler) handleHealthCheckResults(res core.HealthCheckResult) {
	// Healthy > not healthy
	if res.RouterHost.Stats.Healthy && !res.Healthy {
		logrus.Warningf("Router host %v on %v degraded", res.RouterHost.Key(), res.RouterHost.ClusterKey)
	}

	// Not healthy > healthy
	if !res.RouterHost.Stats.Healthy && res.Healthy {
		logrus.Infof("Router host %v on %v became healthy", res.RouterHost.Key(), res.RouterHost.ClusterKey)
	}

	// Update state
	res.RouterHost.Stats.Healthy = res.Healthy
}
