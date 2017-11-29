package balancing

import (
	"errors"
	"math/rand"

	"github.com/ReToCode/openshift-cross-cluster-loadbalancer/balancer/core"
)

func getPossibleRouterHostsBasedOnWeight(hostGroups []*RouterHostGroup) ([]*core.RouterHost, error) {
	totalWeight := 0
	for _, grp := range hostGroups {
		if grp.Weight <= 0 {
			return nil, errors.New("invalid route weight: 0")
		}
		totalWeight += grp.Weight
	}

	r := rand.Intn(totalWeight)
	pos := 0

	for _, grp := range hostGroups {
		pos += grp.Weight
		if r >= pos {
			continue
		}
		return grp.RouterHosts, nil
	}

	return nil, errors.New("error selection router host group based on weight")
}