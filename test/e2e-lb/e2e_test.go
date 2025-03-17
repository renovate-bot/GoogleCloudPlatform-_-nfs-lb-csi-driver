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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/nfs-lb-csi-driver/test/e2e-lb/testsuites"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	storageframework "k8s.io/kubernetes/test/e2e/storage/framework"
)

var _ = func() bool {
	testing.Init()
	if os.Getenv(clientcmd.RecommendedConfigPathEnvVar) == "" {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, kubeconfig)
	}

	framework.RegisterCommonFlags(flag.CommandLine)
	framework.RegisterClusterFlags(flag.CommandLine)
	flag.Parse()
	framework.AfterReadingAllFlags(&framework.TestContext)
	return true
}()

func TestE2E(t *testing.T) {
	t.Parallel()
	gomega.RegisterFailHandler(framework.Fail)
	if framework.TestContext.ReportDir != "" {
		if err := os.MkdirAll(framework.TestContext.ReportDir, 0o755); err != nil {
			klog.Errorf("Failed creating report directory: %v", err)
		}
	}

	suiteConfig, reporterConfig := framework.CreateGinkgoConfig()
	klog.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, suiteConfig.ParallelProcess)
	ginkgo.RunSpecs(t, "NFS LB CSI Driver", suiteConfig, reporterConfig)
}

var _ = ginkgo.Describe("E2E Test Suite", func() {
	CSITestSuites := []func() storageframework.TestSuite{
		testsuites.InitNFSLBCSIVolumesTestSuite,
	}

	testDriver := InitNFSLBCSITestDriver("", true)

	ginkgo.Context(fmt.Sprintf("[Driver: %s]", testDriver.GetDriverInfo().Name), func() {
		storageframework.DefineTestSuites(testDriver, CSITestSuites)
	})
})
