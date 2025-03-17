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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func NewFakeLBController(ipMap map[string]int, nodes []runtime.Object) *LBController {
	client := fake.NewSimpleClientset(nodes...)
	factory := informers.NewSharedInformerFactory(client, time.Hour /* disable resync*/)
	nodeInformer := factory.Core().V1().Nodes()

	for _, obj := range nodes {
		switch obj.(type) {
		case *v1.Node:
			nodeInformer.Informer().GetStore().Add(obj)
		default:
			break
		}
	}

	return &LBController{
		ipMap:      ipMap,
		clientset:  client,
		nodeLister: nodeInformer.Lister(),
	}
}

type TestNode struct {
	Name       string
	AssignedIP string
}

func NewNode(name, assignedIP string) *v1.Node {
	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if assignedIP != "" {
		node.ObjectMeta.Annotations = map[string]string{NodeAnnotation: assignedIP}
	}

	return &node
}

func NewNodePool(fakeNodes []TestNode) []runtime.Object {
	var nodePool []runtime.Object
	for _, fn := range fakeNodes {
		node := NewNode(fn.Name, fn.AssignedIP)
		nodePool = append(nodePool, node)
	}

	return nodePool
}
