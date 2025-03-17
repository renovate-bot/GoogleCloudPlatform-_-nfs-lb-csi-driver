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

	"github.com/GoogleCloudPlatform/nfs-lb-csi-driver/test/e2e-lb/utils"
	"k8s.io/klog/v2"
)

var (
	pkgDir = flag.String("pkg-dir", "", "the package directory")

	// Ginkgo flags.
	ginkgoFocus         = flag.String("ginkgo-focus", "", "pass to ginkgo run --focus flag")
	ginkgoSkip          = flag.String("ginkgo-skip", "", "pass to ginkgo run --skip flag")
	ginkgoProcs         = flag.String("ginkgo-procs", "1", "pass to ginkgo run --procs flag")
	ginkgoTimeout       = flag.String("ginkgo-timeout", "2h", "pass to ginkgo run --timeout flag")
	ginkgoFlakeAttempts = flag.String("ginkgo-flake-attempts", "1", "pass to ginkgo run --flake-attempts flag")
	ginkgoSkipGcpSaTest = flag.Bool("ginkgo-skip-gcp-sa-test", true, "skip GCP SA test")

	// nfs server flags
	nfsServerCount   = flag.Int("nfs-server-count", 0, "if non-zero setup given number of nfs servers")
	gcpProject       = flag.String("gcp-project", "", "GCP project")
	installCSIDriver = flag.Bool("install-csi-driver", false, "if true, install the NFS LB CSI driver")
	destroyCSIDriver = flag.Bool("destroy-csi-driver", false, "if true, destroy the NFS LB CSI driver")
)

func main() {
	flag.Parse()

	testParams := &utils.TestParameters{
		PkgDir:              *pkgDir,
		GinkgoFocus:         *ginkgoFocus,
		GinkgoSkip:          *ginkgoSkip,
		GinkgoProcs:         *ginkgoProcs,
		GinkgoTimeout:       *ginkgoTimeout,
		GinkgoFlakeAttempts: *ginkgoFlakeAttempts,

		NFSServerCount:   *nfsServerCount,
		GcpProject:       *gcpProject,
		InstallCSIDriver: *installCSIDriver,
		DestroyCSIDriver: *destroyCSIDriver,
	}

	if err := utils.Handle(testParams); err != nil {
		klog.Fatalf("Failed to run e2e test: %v", err)
	}
}
