package balancer

import "errors"

type LeastConnectionsBalancer struct {}

func (b *LeastConnectionsBalancer) GetRouterHost(ctx Context, routerHosts []*RouterHost) (*RouterHost, error) {
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

