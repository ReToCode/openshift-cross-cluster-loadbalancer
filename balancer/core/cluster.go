package core

import "github.com/sirupsen/logrus"

type Route struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

type Cluster struct {
	Key         string
	RouterHosts map[string]*RouterHost
	Routes      map[string]Route
}

type ClusterUpdate struct {
	Routes      map[string]Route  `json:"routes"`
	RouterHosts map[string]RouterHost `json:"routerHosts"`
}

func NewCluster(key string, routes map[string]Route) *Cluster {
	// Verify routes
	for _, r := range routes {
		if len(r.URL) == 0 || r.Weight <= 0 {
			logrus.Error("Invalid cluster config!")
		}
	}

	return &Cluster{
		Key:         key,
		RouterHosts: map[string]*RouterHost{},
		Routes:      routes,
	}
}

func (c *Cluster) Stop() {
	for _, rh := range c.RouterHosts {
		rh.Stop()
	}
}
