/*
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lbcontroller

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

const (
	NodeAnnotation = "nfs.lb.csi.storage.gke.io/assigned-ip"
)

type LBController struct {
	clientset  kubernetes.Interface
	nodeLister listersv1.NodeLister
	ipMap      map[string]int
	mutex      sync.Mutex
}

func NewLBController(ipList []string) *LBController {
	klog.Infof("Building kube configs for running in cluster...")
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to create config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}

	ctx := signals.SetupSignalHandler()
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, 10*time.Minute /*Resync interval of the informer*/)
	nodeLister := sharedInformerFactory.Core().V1().Nodes().Lister()
	stopCh := ctx.Done()
	sharedInformerFactory.Start(stopCh)
	sharedInformerFactory.WaitForCacheSync(stopCh)

	lbc := LBController{
		clientset:  clientset,
		nodeLister: nodeLister,
	}

	ipMap, err := lbc.resyncIPMap(ipList)

	if err != nil {
		klog.Fatalf("Failed to resync LB Controller cache: %v", err)
	}

	lbc.ipMap = ipMap

	return &lbc
}

func (c *LBController) resyncIPMap(ipList []string) (map[string]int, error) {
	clusterNodes, err := c.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	ipMap := make(map[string]int)
	for _, ip := range ipList {
		ipMap[ip] = 0
	}

	for _, node := range clusterNodes {
		if ip, exists := node.Annotations[NodeAnnotation]; exists {
			klog.Infof("Node %q already have IP %q assigned", node.Name, ip)
			if _, exists := ipMap[ip]; exists {
				ipMap[ip]++
			}
		}
	}

	klog.V(6).Infof("LB controller ipMap resynced: %v", ipMap)
	return ipMap, nil
}

func (c *LBController) AssignIPToNode(ctx context.Context, nodeName, volumeID string) (string, error) {
	node, err := c.nodeLister.Get(nodeName)
	if err != nil {
		return "", err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ip, exists := node.Annotations[NodeAnnotation]; exists {
		klog.Infof("Node %q already have IP %q assigned", node.Name, ip)
		if _, exists := c.ipMap[ip]; exists {
			return ip, nil
		}
		klog.V(5).Infof("IP %q not found among the NFS server IP list. Reassigning a new IP to node %q", ip, node.Name)
	}

	// Sort IPs by their current count.
	ips := make([]string, 0, len(c.ipMap))
	for ip := range c.ipMap {
		ips = append(ips, ip)
	}
	sort.Slice(ips, func(i, j int) bool {
		return c.ipMap[ips[i]] < c.ipMap[ips[j]]
	})
	selectedIP := ips[0]

	nodeCopy := node.DeepCopy()
	if node.Annotations == nil {
		nodeCopy.Annotations = make(map[string]string)
	}
	nodeCopy.Annotations[NodeAnnotation] = selectedIP

	klog.V(5).Infof("Assigning IP %q to node %q for volume %q", selectedIP, node.Name, volumeID)

	_, err = c.clientset.CoreV1().Nodes().Update(ctx, nodeCopy, metav1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to assign IP %q to node %q: %v", selectedIP, node.Name, err)
	}

	c.ipMap[selectedIP]++
	klog.V(6).Infof("AssignIPToNode: For volume %q, node %q, IP updated %q, LB controller IP map %v", volumeID, nodeName, selectedIP, c.ipMap)
	return selectedIP, nil
}

func (c *LBController) RemoveIPFromNode(ctx context.Context, nodeName, volumeID string) error {
	node, err := c.nodeLister.Get(nodeName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(5).Infof("Node %q not found, skip RemoveIPFromNode for volume %q", nodeName, volumeID)
			return nil
		}
		return err
	}

	ip, ok := node.Annotations[NodeAnnotation]
	if !ok {
		klog.V(5).Infof("Node %q does not have annotation %q, skip RemoveIPFromNode for volume %q", nodeName, NodeAnnotation, volumeID)
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.ipMap[ip]; !exists {
		klog.V(5).Infof("%q does not exist in LB controller IP map, skip RemoveIPFromNode for volume %q", ip, volumeID)
		return nil
	}

	klog.V(5).Infof("Removing IP annotation %q from node %q for volume %q", ip, nodeName, volumeID)
	nodeCopy := node.DeepCopy()
	delete(nodeCopy.Annotations, NodeAnnotation)
	_, err = c.clientset.CoreV1().Nodes().Update(ctx, nodeCopy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	c.ipMap[ip]--
	klog.V(6).Infof("RemoveIPFromNode: For volume %q, node %q, IP updated %q, LB controller IP map %v", volumeID, nodeName, ip, c.ipMap)
	return nil
}
