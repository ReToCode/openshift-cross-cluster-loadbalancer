package stats

import (
	"sync"
	"time"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/sirupsen/logrus"
)

type SafeStats struct {
	v   core.GlobalStats
	mux sync.Mutex
}

type StatsHandler struct {
	// State
	stats           SafeStats
	lastConnections uint

	// Async communication
	Connections    chan uint
	RouterHosts    chan []core.RouterHost
	StatsTick      chan core.GlobalStats
	stop           chan bool
}

func NewHandler() *StatsHandler {
	return &StatsHandler{
		stats: SafeStats{v: core.GlobalStats{
			Mutation:           "stats",
			HealthyHosts:       []int{},
			UnhealthyHosts:     []int{},
			Hosts:              make(map[string]core.RouterHostWithStats),
			OverallConnections: []uint{},
			Ticks:              []string{},
		}},
		lastConnections: 0,

		Connections:    make(chan uint, 1),
		RouterHosts:    make(chan []core.RouterHost),
		StatsTick:      make(chan core.GlobalStats),
		stop:           make(chan bool),
	}
}

func (s *StatsHandler) Start() {
	logrus.Info("Started StatsHandler")

	UIPushTicker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-UIPushTicker.C:
				s.updateGlobalStats()

			case conn := <-s.Connections:
				s.lastConnections = conn

			case c := <-s.RouterHosts:
				s.updateRouterHosts(c)

			case <-s.stop:
				logrus.Info("Stopped StatsHandler")
				return
			}
		}
	}()
}

func (s *StatsHandler) Stop() {
	s.stop <- true
}

func (s *StatsHandler) updateRouterHosts(rhs []core.RouterHost) {
	logrus.Debug("Got a update of the router host map in StatsHandler")

	s.stats.mux.Lock()
	updated := map[string]core.RouterHostWithStats{}

	for i := range rhs {
		rh := rhs[i]
		oldRH, ok := s.stats.v.Hosts[rh.Key()]

		if ok {
			// if we have this router host, update the health state
			logrus.Debugf("Updating existing router %v", rh.Key())

			oldRH.Stats = append(oldRH.Stats, rh.LastState)

			oldRH = s.updateRouterHostStats(oldRH)

			updated[rh.Key()] = oldRH
		} else {
			// router host is new, create a new entry
			logrus.Debugf("New router %v", rh.Key())

			newRH := core.RouterHostWithStats{
				ClusterKey: rh.ClusterKey,
				HostIP:     rh.HostIP,
				HTTPPort:   rh.HTTPPort,
				HTTPSPort:  rh.HTTPSPort,
				Stats:      []core.HostStats{},
			}

			newRH = s.updateRouterHostStats(newRH)
			newRH.Stats = append(newRH.Stats, rh.LastState)

			updated[rh.Key()] = newRH
		}
	}

	s.stats.v.Hosts = updated
	s.stats.mux.Unlock()
}

func (s *StatsHandler) updateRouterHostStats(rh core.RouterHostWithStats) core.RouterHostWithStats {
	if len(rh.Stats) >= core.MaxTicks {
		rh.Stats = rh.Stats[1:]
	} else {
		for i := 0; i <= core.MaxTicks; i++ {
			rh.Stats = append(rh.Stats, core.HostStats{})
		}
	}

	return rh
}

func (s *StatsHandler) updateGlobalStats() {
	logrus.Debug("Updating global stats and sending them to UI")
	unhealthyHosts := 0
	healthyHosts := 0

	s.stats.mux.Lock()
	defer s.stats.mux.Unlock()

	// Update stats for every router host
	for _, rh := range s.stats.v.Hosts {
		if rh.Stats[len(rh.Stats) -1].Healthy {
			healthyHosts++
		} else {
			unhealthyHosts++
		}
	}

	// sort.Slice(hostList, func(i, j int) bool {
	// 	return hostList[i].Key() < hostList[j].Key()
	// })

	// Create a list of ticks and connections for the UI
	if len(s.stats.v.Ticks) >= core.MaxTicks {
		s.stats.v.Ticks = s.stats.v.Ticks[1:]
		s.stats.v.OverallConnections = s.stats.v.OverallConnections[1:]
		s.stats.v.HealthyHosts = s.stats.v.HealthyHosts[1:]
		s.stats.v.UnhealthyHosts = s.stats.v.UnhealthyHosts[1:]
	} else {
		for i := 0; i <= core.MaxTicks; i++ {
			s.stats.v.Ticks = append(s.stats.v.Ticks, "")
			s.stats.v.OverallConnections = append(s.stats.v.OverallConnections, 0)
			s.stats.v.HealthyHosts = append(s.stats.v.HealthyHosts, 0)
			s.stats.v.UnhealthyHosts = append(s.stats.v.UnhealthyHosts, 0)
		}
	}
	s.stats.v.Ticks = append(s.stats.v.Ticks, "")

	s.stats.v.OverallConnections = append(s.stats.v.OverallConnections, s.lastConnections)
	s.stats.v.HealthyHosts = append(s.stats.v.HealthyHosts, healthyHosts)
	s.stats.v.UnhealthyHosts = append(s.stats.v.UnhealthyHosts, unhealthyHosts)

	// Send the stats to the UI
	s.StatsTick <- s.stats.v
}