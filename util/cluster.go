package util

import (
	"strings"

	"github.com/Jeffail/gabs"
)

type ClusterOption struct {
	ClusterName string
	ClusterKey  string
}

// Returns an array of cluster names from the given tf config
func GetClusterOptions(parsedConfig *gabs.Container) ([]*ClusterOption, error) {
	result := []*ClusterOption{}

	children, err := parsedConfig.S("module").ChildrenMap()
	if err != nil {
		return nil, err
	}

	for key, child := range children {
		if strings.Index(key, "cluster_") == 0 {
			name, ok := child.Path("name").Data().(string)
			if !ok {
				continue
			}
			result = append(result, &ClusterOption{
				ClusterKey:  key,
				ClusterName: name,
			})
		}
	}
	return result, nil
}
