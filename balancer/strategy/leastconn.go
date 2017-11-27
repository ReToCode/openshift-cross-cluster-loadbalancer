package strategy

import (
	"errors"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
)

type LeastConnectionsBalancer struct{}

func (b *LeastConnectionsBalancer) GetRouterHost(ctx core.Context, routerHosts []*core.RouterHost) (*core.RouterHost, error) {
	if len(routerHosts) == 0 {
		return nil, errors.New("no available router hosts found")
	}

	least := routerHosts[0]
	for key, rh := range routerHosts {
		if rh.Stats.ActiveConnections <= least.Stats.ActiveConnections {
			least = routerHosts[key]
		}
	}

	return least, nil
}