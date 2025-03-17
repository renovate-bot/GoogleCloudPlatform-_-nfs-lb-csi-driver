/*
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testsuites

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"text/template"

	"github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	storageframework "k8s.io/kubernetes/test/e2e/storage/framework"
)

const (
	nfsClientDeploymentTemplatePath = "./templates/nfs-client-v3.yaml"
	NodeAnnotation                  = "nfs.lb.csi.storage.gke.io/assigned-ip"
)

type NFSLBCSIVolumesTestSuite struct {
	tsInfo storageframework.TestSuiteInfo
}

type NodeValidationOptions struct {
	ExpectedNodesWithTargetAnnotation int
	ExpectedUniqueIPAddress           int
	ExpectedIPToNodeMinCount          int
	ExpectedIPToNodeMaxCount          int
}

type NFSClientOptions struct {
	ClientName string
	PVName     string
	PVCName    string
	Replicas   int
}

func InitNFSLBCSIVolumesTestSuite() storageframework.TestSuite {
	return &NFSLBCSIVolumesTestSuite{
		tsInfo: storageframework.TestSuiteInfo{
			Name: "volumes",
			TestPatterns: []storageframework.TestPattern{
				// storageframework.DefaultFsCSIEphemeralVolume,
				storageframework.DefaultFsPreprovisionedPV,
				// storageframework.DefaultFsDynamicPV,
			},
		},
	}
}

func (t *NFSLBCSIVolumesTestSuite) GetTestSuiteInfo() storageframework.TestSuiteInfo {
	return t.tsInfo
}

func (t *NFSLBCSIVolumesTestSuite) SkipUnsupportedTests(_ storageframework.TestDriver, _ storageframework.TestPattern) {
}

func (t *NFSLBCSIVolumesTestSuite) DefineTests(driver storageframework.TestDriver, pattern storageframework.TestPattern) {
	// Beware that it also registers an AfterEach which renders f unusable. Any code using
	// f must run inside an It or Context callback.
	f := framework.NewFrameworkWithCustomTimeouts("volumes", storageframework.GetDriverTimeouts(driver))

	// This test case expects 10 GKE nodes, 10 NFS Servers and creates 10 NFS clients (1 on each node).
	ginkgo.It("bringup-teardown-10node-10nfsclient-10nfsserver", func() {
		nfsClientPodReplicas := 10
		expectedNFSServers := 10
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client",
			PVName:     "nfs-client-pv",
			PVCName:    "nfs-client-pvc",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          1,
			ExpectedIPToNodeMaxCount:          1,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client",
				PVName:     "nfs-client-pv",
				PVCName:    "nfs-client-pvc",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}

			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})

	// This test case expects 10 GKE nodes, 5 NFS Servers and creates 5 NFS clients (1 on each node).
	ginkgo.It("bringup-teardown-10node-5nfsclient-5nfsserver", func() {
		nfsClientPodReplicas := 5
		expectedNFSServers := 5
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client",
			PVName:     "nfs-client-pv",
			PVCName:    "nfs-client-pvc",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          1,
			ExpectedIPToNodeMaxCount:          1,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client",
				PVName:     "nfs-client-pv",
				PVCName:    "nfs-client-pvc",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}

			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})

	// This test case expects 10 GKE nodes, 5 NFS Servers and creates 10 NFS clients (1 on each node).
	ginkgo.It("bringup-teardown-10node-10nfsclient-5nfsserver", func() {
		nfsClientPodReplicas := 10
		expectedNFSServers := 5
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client",
			PVName:     "nfs-client-pv",
			PVCName:    "nfs-client-pvc",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          2,
			ExpectedIPToNodeMaxCount:          2,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client",
				PVName:     "nfs-client-pv",
				PVCName:    "nfs-client-pvc",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}

			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})

	// This test case expects 10 GKE nodes, 5 NFS Servers and creates 9 NFS clients (1 on each node).
	ginkgo.It("bringup-teardown-10node-9nfsclient-5nfsserver", func() {
		nfsClientPodReplicas := 9
		expectedNFSServers := 5
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client",
			PVName:     "nfs-client-pv",
			PVCName:    "nfs-client-pvc",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          1,
			ExpectedIPToNodeMaxCount:          2,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client",
				PVName:     "nfs-client-pv",
				PVCName:    "nfs-client-pvc",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}

			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})

	// Multi NFS Client deployment test cases
	// This test case expects 10 GKE nodes, 5 NFS Servers and creates 2 5replica NFS clients (1 on each node).
	ginkgo.It("bringup-teardown-10node-2-5nfsclient-5nfsserver", func() {
		nfsClientPodReplicas := 5
		expectedNFSServers := 5
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client-0",
			PVName:     "nfs-client-pv-0",
			PVCName:    "nfs-client-pvc-0",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		err = setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client-1",
			PVName:     "nfs-client-pv-1",
			PVCName:    "nfs-client-pvc-1",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}
		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: 2 * nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          2,
			ExpectedIPToNodeMaxCount:          2,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client-0",
				PVName:     "nfs-client-pv-0",
				PVCName:    "nfs-client-pvc-0",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}
			err = validateNodeAnnotations(&NodeValidationOptions{
				ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
				ExpectedIPToNodeMinCount:          1,
				ExpectedIPToNodeMaxCount:          1,
				ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
			}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}

			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client-1",
				PVName:     "nfs-client-pv-1",
				PVCName:    "nfs-client-pvc-1",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}
			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})

	ginkgo.It("bringup-teardown-100node-100nfsclient-10nfsserver", func() {
		nfsClientPodReplicas := 100
		expectedNFSServers := 10
		err := setupNFSClient(&NFSClientOptions{
			Replicas:   nfsClientPodReplicas,
			ClientName: "nfs-client",
			PVName:     "nfs-client-pv",
			PVCName:    "nfs-client-pvc",
		}, false /* destroy */)
		if err != nil {
			klog.Fatalf("failed to setup nfs client %v", err)
		}

		nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("failed to list nodes in the cluster %v", err)
		}
		klog.Infof("found %d nodes", len(nodes.Items))
		err = validateNodeAnnotations(&NodeValidationOptions{
			ExpectedNodesWithTargetAnnotation: nfsClientPodReplicas,
			ExpectedIPToNodeMinCount:          10,
			ExpectedIPToNodeMaxCount:          10,
			ExpectedUniqueIPAddress:           expectedNFSServers, // number of nfs servers
		}, nodes.Items)
		if err != nil {
			klog.Fatalf("node validation failed with error %v", err)
		}

		defer func() {
			setupNFSClient(&NFSClientOptions{
				Replicas:   nfsClientPodReplicas,
				ClientName: "nfs-client",
				PVName:     "nfs-client-pv",
				PVCName:    "nfs-client-pvc",
			}, true /* destroy */)
			nodes, err = f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				klog.Fatalf("failed to list nodes in the cluster %v", err)
			}

			err = validateNodeAnnotations(&NodeValidationOptions{}, nodes.Items)
			if err != nil {
				klog.Fatalf("node validation failed with error %v", err)
			}
		}()
	})
}

func validateNodeAnnotations(c *NodeValidationOptions, nodes []v1.Node) error {
	numNodesWithAnn := 0
	ipMap := make(map[string]int)
	for _, n := range nodes {
		if ip, ok := n.Annotations[NodeAnnotation]; ok {
			numNodesWithAnn = numNodesWithAnn + 1
			ipMap[ip]++
		}
	}

	klog.Infof("IPMap %v", ipMap)
	if numNodesWithAnn != c.ExpectedNodesWithTargetAnnotation {
		return fmt.Errorf("Mismatch in expected nodes with annotation, got %d, expected %d", numNodesWithAnn, c.ExpectedNodesWithTargetAnnotation)
	}

	if numNodesWithAnn == 0 {
		return nil
	}

	uniqueIPAddr := len(ipMap)
	if uniqueIPAddr != c.ExpectedUniqueIPAddress {
		return fmt.Errorf("Mismatch in unique IP Addresses got %d, expected %d", uniqueIPAddr, c.ExpectedUniqueIPAddress)
	}

	minNodeCount := 0
	maxNodeCount := 0
	first := true
	for _, v := range ipMap {
		if first {
			minNodeCount = v
			maxNodeCount = v
			first = false
		}

		if v < minNodeCount {
			minNodeCount = v
		}
		if v > maxNodeCount {
			maxNodeCount = v
		}
	}

	if minNodeCount != c.ExpectedIPToNodeMinCount {
		return fmt.Errorf("Mismatch in IP Address to node min count got %d, expected %d", minNodeCount, c.ExpectedIPToNodeMinCount)
	}
	if maxNodeCount != c.ExpectedIPToNodeMaxCount {
		return fmt.Errorf("Mismatch in IP Address to node max count got %d, expected %d", maxNodeCount, c.ExpectedIPToNodeMaxCount)
	}
	return nil
}

func setupNFSClient(o *NFSClientOptions, destroy bool) error {
	operation := "apply"
	if destroy {
		operation = "delete"
	}
	klog.Infof("Starting operation %q for %d nfs client replicas", operation, o.Replicas)
	templateData, err := ioutil.ReadFile(nfsClientDeploymentTemplatePath)
	if err != nil {
		return fmt.Errorf("error reading template file %v: %w", nfsClientDeploymentTemplatePath, err)
	}

	values := struct {
		NFSClientName    string
		NFSClientPVName  string
		NFSClientPVCName string
		Replicas         string
	}{
		NFSClientName:    o.ClientName,
		NFSClientPVName:  o.PVName,
		NFSClientPVCName: o.PVCName,
		Replicas:         strconv.Itoa(o.Replicas),
	}

	tmpl, err := template.New("deployment").Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("error parsing template:%w", err)
	}

	var populatedTemplate bytes.Buffer
	err = tmpl.Execute(&populatedTemplate, values)
	if err != nil {
		return fmt.Errorf("error executing template:%w", err)
	}

	tmpFile, err := ioutil.TempFile("", "populated-*.yaml")
	if err != nil {
		return fmt.Errorf("error creating temporary file:%w", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(populatedTemplate.Bytes())
	if err != nil {
		return fmt.Errorf("error writing to temporary file:%w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("kubectl", operation, "-f", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	klog.Info(string(output))
	if err != nil {
		return fmt.Errorf("error executing kubectl:%w", err)
	}

	if destroy {
		return nil
	}

	klog.Info("wait for deployment to be ready")
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "deployment/"+o.ClientName, "--timeout=10m")
	output, err = cmd.CombinedOutput()
	klog.Info(string(output))
	if err != nil {
		return fmt.Errorf("error executing kubectl wait:%w", err)
	}
	klog.Info("deployment is ready")
	return nil
}
