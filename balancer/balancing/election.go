package balancing

import (
	"errors"

	"strings"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
	"github.com/sirupsen/logrus"
)

type RouterHostGroup struct {
	Weight      int
	RouterHosts []*core.RouterHost
}

func ElectRouterHost(ctx core.Context, clusters map[string]*core.Cluster) (*core.RouterHost, error) {
	if len(clusters) == 0 {
		return nil, errors.New("can't elect router host, no OpenShift cluster defined")
	}

	var possibleRouterHosts []*core.RouterHost

	if len(ctx.Hostname) > 0 {
		var hostGroups []*RouterHostGroup

		// Check if cluster does handle that route
		for _, cl := range clusters {
			for _, r := range cl.Routes {
				if strings.ToLower(strings.TrimSpace(r.URL)) == strings.ToLower(strings.TrimSpace(ctx.Hostname)) {
					grp := &RouterHostGroup{
						RouterHosts: []*core.RouterHost{},
						Weight:      r.Weight,
					}

					// Add every healthy router of that cluster
					for _, rh := range cl.RouterHosts {
						if !rh.LastState.Healthy {
							continue
						}
						grp.RouterHosts = append(grp.RouterHosts, rh)
					}

					hostGroups = append(hostGroups, grp)

					// Routes are unique
					break
				}
			}
		}

		// Check if route was found on any cluster
		if len(hostGroups) > 0 {
			var err error
			possibleRouterHosts, err = getPossibleRouterHostsBasedOnWeight(hostGroups)
			if err != nil {
				logrus.Error(err.Error())
			}
		} else {
			logrus.Warnf("Route '%v' has no valid target router hosts on any cluster. Balancing to all healthy router hosts", ctx.Hostname)
		}
	} else {
		logrus.Warnf("No route name was parsed. Balancing to all healthy router hosts")
	}

	if len(possibleRouterHosts) == 0 {
		for _, cl := range clusters {
			for _, rh := range cl.RouterHosts {
				if !rh.LastState.Healthy {
					continue
				}
				possibleRouterHosts = append(possibleRouterHosts, rh)
			}
		}
	}

	// From all possible router hosts get the one with the least connections
	return getRouterHostWithLeastConn(possibleRouterHosts)
}
