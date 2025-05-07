package ui

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func adaptRouteConfigsToUI(configs []*domain.RouteConfiguration, vhVersions map[int32][]*domain.DeploymentVersion) []SimplifiedRouteConfig {
	result := make([]SimplifiedRouteConfig, len(configs))
	for i, config := range configs {
		result[i] = adaptRouteConfigToUI(config, vhVersions)
	}
	return result
}

func adaptRouteConfigToUI(config *domain.RouteConfiguration, vhVersions map[int32][]*domain.DeploymentVersion) SimplifiedRouteConfig {
	return SimplifiedRouteConfig{
		NodeGroup:    config.NodeGroupId,
		VirtualHosts: adaptVirtualHostsToUI(config.VirtualHosts, vhVersions),
	}
}

func adaptVirtualHostsToUI(hosts []*domain.VirtualHost, versions map[int32][]*domain.DeploymentVersion) []VirtualHost {
	result := make([]VirtualHost, len(hosts))
	for i, vh := range hosts {
		result[i] = adaptVirtualHostToUI(vh, versions)
	}
	return result
}

func adaptVirtualHostToUI(vh *domain.VirtualHost, versions map[int32][]*domain.DeploymentVersion) VirtualHost {
	return VirtualHost{
		Id:                 vh.Id,
		Name:               vh.Name,
		DeploymentVersions: adaptDeploymentVersions(versions[vh.Id]),
	}
}

func adaptDeploymentVersions(versions []*domain.DeploymentVersion) []DeploymentVersion {
	result := make([]DeploymentVersion, len(versions))
	for i, version := range versions {
		result[i] = DeploymentVersion{
			Name:  version.Version,
			Stage: version.Stage,
		}
	}
	return result
}

func adaptRoutesToUI(routes []*domain.Route) []Route {
	result := make([]Route, len(routes))
	for i, route := range routes {
		result[i] = Route{
			Uuid:        route.Uuid,
			ClusterName: route.ClusterName,
			Allowed:     route.DirectResponseCode != 404,
		}
		if route.Prefix != "" {
			result[i].Match = route.Prefix
			result[i].MatchRewrite = route.PrefixRewrite
			continue
		}
		if route.Path != "" {
			result[i].Match = route.Path
			result[i].MatchRewrite = route.PathRewrite
			continue
		}
		if route.Regexp != "" {
			result[i].Match = route.Regexp
			result[i].MatchRewrite = route.RegexpRewrite
		}
	}
	return result
}

func adaptClusterToUI(clusters []*domain.Cluster, versions []*domain.DeploymentVersion) []Cluster {
	result := make([]Cluster, len(clusters))
	for i, cluster := range clusters {
		result[i] = Cluster{
			Id:                 cluster.Id,
			Name:               cluster.Name,
			EnableH2:           cluster.EnableH2,
			NodeGroups:         adaptNodeGroupsToUI(cluster.NodeGroups),
			DeploymentVersions: adaptDeploymentVersionsExt(versions, cluster.Endpoints),
		}
	}
	return result
}

func adaptNodeGroupsToUI(groups []*domain.NodeGroup) []NodeGroup {
	result := make([]NodeGroup, len(groups))
	for i, nodeGroup := range groups {
		result[i] = NodeGroup{
			nodeGroup.Name,
		}
	}
	return result
}

func adaptDeploymentVersionsExt(versions []*domain.DeploymentVersion, endpoints []*domain.Endpoint) []DeploymentVersionsExt {
	versionsMap := make(map[string]*domain.DeploymentVersion)
	for _, version := range versions {
		versionsMap[version.Version] = version
	}
	result := make([]DeploymentVersionsExt, 0)
	// 1 endpoint -> 1 version
	for _, endpoint := range endpoints {
		dVersion := versionsMap[endpoint.DeploymentVersion]
		result = append(result, DeploymentVersionsExt{
			Name:  dVersion.Version,
			Stage: dVersion.Stage,
			Endpoint: Endpoint{
				Id:      endpoint.Id,
				Address: endpoint.Address,
				Port:    endpoint.Port,
			},
			BalancingPolicies: adaptHashPoliciesToUI(endpoint.HashPolicies),
		})
	}
	return result
}

func adaptHashPoliciesToUI(policies []*domain.HashPolicy) []BalancingPolicy {
	result := make([]BalancingPolicy, len(policies))
	for i, policy := range policies {
		result[i] = BalancingPolicy{
			Id: policy.Id,
		}
		if policy.HeaderName != "" {
			result[i].Header = &Header{
				HeaderName: policy.HeaderName,
			}
		}
		if policy.CookieName != "" {
			var ttlValue *int64
			if policy.CookieTTL.Valid {
				ttlValue = &policy.CookieTTL.Int64
			}

			result[i].Cookie = &Cookie{
				Name: policy.CookieName,
				Ttl:  ttlValue,
				Path: domain.NewNullString(policy.CookiePath),
			}
		}
		if policy.QueryParamName != "" {
			result[i].QueryParameter = &QueryParameter{
				Name: policy.QueryParamName,
			}
		}
		if policy.QueryParamSourceIP.Valid {
			result[i].ConnectionProperties = &ConnectionProperties{
				SourceIp: policy.QueryParamSourceIP,
			}

		}
		result[i].Terminal = policy.Terminal
	}
	return result
}
