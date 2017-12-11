package balancing

import (
"errors"

"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
)

func getRouterHostWithLeastConn(routerHosts []*core.RouterHost) (*core.RouterHost, error) {
	if len(routerHosts) == 0 {
		return nil, errors.New("no available router hosts found")
	}

	least := routerHosts[0]
	for key, rh := range routerHosts {
		if rh.LastState.ActiveConnections <= least.LastState.ActiveConnections {
			least = routerHosts[key]
		}
	}

	return least, nil
}
