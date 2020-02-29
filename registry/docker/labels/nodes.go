package dockerLabels

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

type nodeCheckMethod func([]swarm.Node, int) bool

func isStateReady(node []swarm.Node, current int) bool {
	return node[current].Status.State == swarm.NodeStateReady
}
func isAvailabilityActive(node []swarm.Node, current int) bool {
	return node[current].Spec.Availability == swarm.NodeAvailabilityActive
}
func isWorker(node []swarm.Node, current int) bool {
	return node[current].Spec.Role == swarm.NodeRoleWorker
}
func isMulti(node []swarm.Node, current int) bool {
	return len(node) > 1
}

func Normal(aa ...nodeCheckMethod) nodeCheckMethod {
	return func(nodes []swarm.Node, i int) bool {
		return isStateReady(nodes, i) && isAvailabilityActive(nodes, i) && And(aa...)(nodes, i)
	}
}

func Not(f nodeCheckMethod) nodeCheckMethod {
	return func(nodes []swarm.Node, i int) bool {
		return !f(nodes, i)
	}
}

func Or(aa ...nodeCheckMethod) nodeCheckMethod {
	return func(nodes []swarm.Node, i int) bool {
		for _, f := range aa {
			if f(nodes, i) {
				return true
			}
		}
		return false
	}
}

func And(aa ...nodeCheckMethod) nodeCheckMethod {
	return func(nodes []swarm.Node, i int) bool {
		for _, f := range aa {
			if !f(nodes, i) {
				return false
			}
		}
		return true
	}
}

func (dr *DockerLabelsRegister) getNodes(filters ...nodeCheckMethod) ([]string, error) {
	if nodes, err := dr.docker.NodeList(context.TODO(), types.NodeListOptions{}); err != nil {
		return nil, err
	} else {
		nodeIps := make([]string, 0)
	NODES:
		for current, node := range nodes {
			for _, filter := range filters {
				if !filter(nodes, current) {
					continue NODES
				}
			}
			nodeIps = append(nodeIps, node.Status.Addr)
		}
		return nodeIps, nil
	}
}
