package main

import (
	"github.com/pkg/errors"
	"gopkg.in/gcfg.v1"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud"
	provider_os "k8s.io/kubernetes/pkg/cloudprovider/providers/openstack"
	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/stackresources"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type Watcher struct {
	clusterClient *gophercloud.ServiceClient
	heatClient    *gophercloud.ServiceClient

	kubeClient *kubernetes.Clientset

	stackID string
	stackName string
}

func CreateWatcher() (*Watcher, error) {
	var cfg provider_os.Config
	if err := gcfg.ReadFileInto(&cfg, *cloudConfig); err != nil {
		return nil, errors.Wrap(err, "could not read cloud config")
	}

	authOpts := toAuthOptsExt(cfg)

	provider, err := openstack.NewClient(cfg.Global.AuthURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not authenticate client")
	}

	err = openstack.AuthenticateV3(provider, authOpts, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, errors.Wrap(err, "could not authenticate")
	}

	clusterClient, err := openstack.NewContainerInfraV1(provider, gophercloud.EndpointOpts{Type: "container-infra", Name: "magnum", Region: cfg.Global.Region})
	if err != nil {
		return nil, errors.Wrap(err, "could not create container-infra client")
	}

	heatClient, err := openstack.NewOrchestrationV1(provider, gophercloud.EndpointOpts{Type: "orchestration", Name: "heat", Region: cfg.Global.Region})
	if err != nil {
		return nil, errors.Wrap(err, "could not create orchestration client")
	}

	kubeClient, err := CreateKubeClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not create kubeClient")
	}

	watcher := &Watcher{
		clusterClient: clusterClient,
		heatClient: heatClient,
		kubeClient: kubeClient,
	}

	watcher.stackID, err = GetStackID(clusterClient, *clusterName)
	if err != nil {
		return nil, errors.Wrap(err, "could not set stack ID")
	}

	watcher.stackName, err = GetStackName(heatClient, watcher.stackID)
	if err != nil {
		return nil, errors.Wrap(err, "could not set stack name")
	}

	return watcher, nil
}

func (w *Watcher) GetRemovedMinions() ([]string, error) {
	minions, err := stackresources.Get(w.heatClient, w.stackName, w.stackID, "kube_minions").Extract()
	if err != nil {
		return nil, errors.Wrap(err, "could not get kube_minions")
	}
	// removed_rsrc_list []interface{}{"1", "2", "3", ...}
	removedMinions := minions.Attributes["removed_rsrc_list"].([]interface{})
	var removedMinionNums []string
	for _, m := range removedMinions {
		removedMinionNums = append(removedMinionNums, m.(string))
	}
	return removedMinionNums, nil
}

func (w *Watcher) Tick() error {
	removedMinions, err := w.GetRemovedMinions()
	if err != nil {
		return err
	}

	// glog.Infof("Ticking, removedMinions=%v", removedMinions)

	nodeList, err := w.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "could not list nodes")
	}

	for _, node := range nodeList.Items {
		// Never delete a ready node
		nodeReady := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == v1.NodeReady {
				if cond.Status == v1.ConditionTrue {
					nodeReady = true
				}
			}
		}
		if nodeReady {
			continue
		}

		glog.Infof("Checking node %s, shouldBeDeleted=%t", node.Name, NodeShouldBeDeleted(node.Name, removedMinions))
		if NodeShouldBeDeleted(node.Name, removedMinions) {
			glog.Infof("Deleting node: %s", node.Name)
			err := w.kubeClient.CoreV1().Nodes().Delete(node.Name, &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "could not delete node %s", node.Name)
			}
		}
	}

	return nil
}