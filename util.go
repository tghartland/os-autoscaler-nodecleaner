package main

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/containerinfra/v1/clusters"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/trusts"
	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	provider_os "k8s.io/kubernetes/pkg/cloudprovider/providers/openstack"
	"strings"
)

func CreateKubeClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "could not get in cluster config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, nil
}

func toAuthOptsExt(cfg provider_os.Config) trusts.AuthOptsExt {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.Global.AuthURL,
		Username:         cfg.Global.Username,
		UserID:           cfg.Global.UserID,
		Password:         cfg.Global.Password,
		TenantID:         cfg.Global.TenantID,
		TenantName:       cfg.Global.TenantName,
		DomainID:         cfg.Global.DomainID,
		DomainName:       cfg.Global.DomainName,

		// Persistent service, so we need to be able to renew tokens.
		AllowReauth: true,
	}

	return trusts.AuthOptsExt{
		TrustID:            cfg.Global.TrustID,
		AuthOptionsBuilder: &opts,
	}
}

func GetStackID(clusterClient *gophercloud.ServiceClient, clusterName string) (string, error) {
	cluster, err := clusters.Get(clusterClient, clusterName).Extract()
	if err != nil {
		return "", errors.Wrap(err, "could not get cluster")
	}
	return cluster.StackID, nil
}

func GetStackName(heatClient *gophercloud.ServiceClient, stackID string) (string, error) {
	stack, err := stacks.Get(heatClient, "", stackID).Extract()
	if err != nil {
		return "", errors.Wrap(err, "could not get stack")
	}
	return stack.Name, nil
}

func NodeShouldBeDeleted(nodeName string, removedMinionNums []string) bool {
	nodeNameParts := strings.Split(nodeName, "-")
	nodeNum := nodeNameParts[len(nodeNameParts)-1]
	for _, removedMinionNum := range removedMinionNums {
		if nodeNum == removedMinionNum {
			return true
		}
	}
	return false
}
