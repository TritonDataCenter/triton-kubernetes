package util

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
)

type NodeOption struct {
	NodeKey  string
	Hostname string
}

// Returns nodes associated with the given cluster key
func GetNodeOptions(parsedConfig *gabs.Container, clusterKey string) ([]*NodeOption, error) {
	result := []*NodeOption{}

	children, err := parsedConfig.S("module").ChildrenMap()
	if err != nil {
		return nil, err
	}

	for key, child := range children {
		if strings.Index(key, "node_") == 0 {
			// Ignoring node if it doesn't belong to the cluster
			envID, ok := child.Path("rancher_environment_id").Data().(string)
			if !ok || !strings.Contains(envID, fmt.Sprintf(".%s.", clusterKey)) {
				continue
			}

			// Retrieving hostname
			hostname, ok := child.Path("hostname").Data().(string)
			if !ok {
				continue
			}

			result = append(result, &NodeOption{
				NodeKey:  key,
				Hostname: hostname,
			})
		}
	}
	return result, nil
}
