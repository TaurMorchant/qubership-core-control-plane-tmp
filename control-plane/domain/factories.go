package domain

import "time"

func NewNodeGroup(name string) *NodeGroup {
	return &NodeGroup{
		Name: name,
	}
}

func NewClusterNodeGroups(clusterId int32, nodeGroupName string) *ClustersNodeGroup {
	return &ClustersNodeGroup{
		ClustersId:     clusterId,
		NodegroupsName: nodeGroupName,
	}
}

func NewEndpoint(address string, port int32, deploymentVersion, initialDeploymentVersion string, clusterId int32) *Endpoint {
	return &Endpoint{
		Address:                  address,
		Port:                     port,
		DeploymentVersion:        deploymentVersion,
		InitialDeploymentVersion: initialDeploymentVersion,
		ClusterId:                clusterId,
	}
}

func NewDeploymentVersion(version string, stage string) *DeploymentVersion {
	return &DeploymentVersion{
		Version:     version,
		Stage:       stage,
		CreatedWhen: time.Now(),
		UpdatedWhen: time.Now(),
	}
}

func NewListener(listenerName string, bindHost string, bindPort string,
	nodeGroupName string, routeConfigurationName string) *Listener {
	return &Listener{
		Name:                   listenerName,
		BindHost:               bindHost,
		BindPort:               bindPort,
		NodeGroupId:            nodeGroupName,
		RouteConfigurationName: routeConfigurationName,
		Version:                1,
	}
}

func NewCluster(name string, enableH2 bool) *Cluster {
	if enableH2 {
		var httpVersion int32 = 2
		NewCluster2(name, &httpVersion)
	}
	return NewCluster2(name, nil)
}

func NewCluster2(name string, httpVersion *int32) *Cluster {
	return &Cluster{
		Name:             name,
		LbPolicy:         LbPolicyLeastRequest,
		DiscoveryType:    DISCOVERY_TYPE_STRICT_DNS,
		DiscoveryTypeOld: DISCOVERY_TYPE_STRICT_DNS,
		HttpVersion:      httpVersion,
		EnableH2:         httpVersion != nil && *httpVersion == 2,
		Version:          1,
	}
}

func NewRouteConfiguration(routeConfigName string, nodeGroupName string) *RouteConfiguration {
	return &RouteConfiguration{
		Name:        routeConfigName,
		NodeGroupId: nodeGroupName,
		Version:     1,
	}
}

func NewVirtualHost(virtualHostName string, routeConfigurationId int32) *VirtualHost {
	return &VirtualHost{
		Name:                 virtualHostName,
		RouteConfigurationId: routeConfigurationId,
		Version:              1,
	}
}

func NewVirtualHostDomain(domain string, virtualHostId int32) *VirtualHostDomain {
	return &VirtualHostDomain{
		Domain:        domain,
		Version:       1,
		VirtualHostId: virtualHostId,
	}
}

func NewEnvoyConfigVersion(nodeGroup, entityType string) *EnvoyConfigVersion {
	timeInNanoSec := time.Now().UnixNano()
	return &EnvoyConfigVersion{
		NodeGroup:  nodeGroup,
		EntityType: entityType,
		Version:    timeInNanoSec,
	}
}
