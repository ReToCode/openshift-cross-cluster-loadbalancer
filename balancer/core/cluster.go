package core

import "github.com/sirupsen/logrus"

type Route struct {
	URL    string
	Weight int
}

type Cluster struct {
	Key         string
	RouterHosts map[string]*RouterHost
	Routes      []Route
}

func NewCluster(key string, routes []Route) *Cluster {
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
