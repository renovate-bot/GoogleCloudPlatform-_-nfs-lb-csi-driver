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

package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"k8s.io/klog/v2"
)

const (
	nfsServerTemplateFilePath = "/test/e2e-lb/templates/nfs-server-v3.yaml"
)

var envAPIMap = map[string]string{
	"https://container.googleapis.com/":                  "prod",
	"https://staging-container.sandbox.googleapis.com/":  "staging",
	"https://staging2-container.sandbox.googleapis.com/": "staging2",
	"https://test-container.sandbox.googleapis.com/":     "test",
}

type TestParameters struct {
	PkgDir              string
	GinkgoSkip          string
	GinkgoFocus         string
	GinkgoProcs         string
	GinkgoTimeout       string
	GinkgoFlakeAttempts string
	GinkgoSkipGcpSaTest bool

	NFSServerCount   int
	GcpProject       string
	InstallCSIDriver bool
	DestroyCSIDriver bool
}

func Handle(testParams *TestParameters) error {
	oldMask := syscall.Umask(0o000)
	defer syscall.Umask(oldMask)
	artifactsDir, ok := os.LookupEnv("ARTIFACTS")
	if !ok {
		artifactsDir = testParams.PkgDir + "/_artifacts"
	}

	testFocusStr := testParams.GinkgoFocus
	if len(testFocusStr) != 0 {
		testFocusStr = fmt.Sprintf(".*%s.*", testFocusStr)
	}

	nfsServerIPList, err := setupNFSServers(testParams.NFSServerCount, testParams.PkgDir+nfsServerTemplateFilePath, false /* unisntall */)
	if err != nil {
		klog.Fatalf("setup of nfs server failed: %v", err)
	}

	klog.Info("nfs server IP list:", nfsServerIPList)
	defer setupNFSServers(testParams.NFSServerCount, testParams.PkgDir+nfsServerTemplateFilePath, true /* unisntall */)

	if testParams.InstallCSIDriver {
		// Install the CSI driver
		err = installDriver(testParams.GcpProject, nfsServerIPList)
		if err != nil {
			klog.Fatalf("setup of CSI driver failed: %v", err)
		}
	}

	defer func() {
		if testParams.DestroyCSIDriver {
			err = deleteDriver(testParams.GcpProject)
			if err != nil {
				klog.Fatalf("destroy of CSI driver failed: %v", err)
			}
		}
	}()

	// nolint:gosec
	cmd := exec.Command("ginkgo", "run", "-v",
		"--procs", testParams.GinkgoProcs,
		"--flake-attempts", testParams.GinkgoFlakeAttempts,
		"--timeout", testParams.GinkgoTimeout,
		"--focus", testFocusStr,
		"--skip", generateTestSkip(testParams),
		"--junit-report", "junit-nfscsi.xml",
		"--output-dir", artifactsDir,
		testParams.PkgDir+"/test/e2e-lb/",
		"--",
		"--provider", "skeleton",
	)

	if err := runCommand("Running Ginkgo e2e test...", cmd); err != nil {
		return fmt.Errorf("failed to run e2e tests with ginkgo: %w", err)
	}

	return nil
}

func generateTestSkip(testParams *TestParameters) string {
	skipTests := []string{}

	if testParams.GinkgoSkip != "" {
		skipTests = append(skipTests, testParams.GinkgoSkip)
	}

	skipString := strings.Join(skipTests, "|")
	klog.Infof("Generated ginkgo skip string: %q", skipString)
	return skipString
}

func setupNFSServers(count int, templateFilePath string, uninstall bool) ([]string, error) {
	operation := "apply"
	if uninstall {
		operation = "delete"
	}
	klog.Infof("Starting operation %q for %d nfs servers", operation, count)

	templateData, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return []string{}, fmt.Errorf("error reading template file %v: %w", templateFilePath, err)
	}

	for i := 0; i < count; i++ {
		values := struct {
			NFSServerName    string
			NFSServerPVCName string
		}{
			NFSServerName:    "nfs-server-" + strconv.Itoa(i),
			NFSServerPVCName: "nfs-server-pvc-" + strconv.Itoa(i),
		}

		tmpl, err := template.New("deployment").Parse(string(templateData))
		if err != nil {
			return []string{}, fmt.Errorf("error parsing template: %w", err)
		}

		var populatedTemplate bytes.Buffer
		err = tmpl.Execute(&populatedTemplate, values)
		if err != nil {
			return []string{}, fmt.Errorf("error executing template: %w", err)
		}

		tmpFile, err := ioutil.TempFile("", "populated-*.yaml")
		if err != nil {
			return []string{}, fmt.Errorf("error creating temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(populatedTemplate.Bytes())
		if err != nil {
			return []string{}, fmt.Errorf("error writing to temporary file: %w", err)
		}
		tmpFile.Close()

		cmd := exec.Command("kubectl", operation, "-f", tmpFile.Name())
		output, err := cmd.CombinedOutput()
		klog.Info(string(output))
		if err != nil {
			return []string{}, fmt.Errorf("error executing kubectl: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	if uninstall {
		return []string{}, nil
	}

	klog.Infof("Wait for %d deployments to be ready", count)
	nfsServerips := []string{}
	for i := 0; i < count; i++ {
		cmd := exec.Command("kubectl", "wait", "--for=condition=available", "deployment/nfs-server-"+strconv.Itoa(i))
		output, err := cmd.CombinedOutput()
		klog.Info(string(output))
		if err != nil {
			return []string{}, fmt.Errorf("error executing kubectl wait: %w", err)
		}

		cmd = exec.Command("kubectl", "get", "service", "nfs-server-"+strconv.Itoa(i), "-o", "jsonpath={.spec.clusterIP}")
		output, err = cmd.CombinedOutput()
		klog.Infof("nfs server ip %q for deployment %d", string(output), i)
		if err != nil {
			return []string{}, fmt.Errorf("error executing kubectl get service: %w", err)
		}
		ip := strings.TrimSpace(string(output))
		nfsServerips = append(nfsServerips, ip)
		time.Sleep(1 * time.Second)
	}

	return nfsServerips, nil
}
