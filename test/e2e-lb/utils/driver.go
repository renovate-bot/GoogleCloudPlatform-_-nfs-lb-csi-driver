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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"k8s.io/klog/v2"
)

func installDriver(project string, ipAddressList []string) error {
	tmpFile, err := ioutil.TempFile("", "ip_addresses.txt")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	writer := bufio.NewWriter(tmpFile)
	for _, ip := range ipAddressList {
		_, err := fmt.Fprintln(writer, ip)
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}
	writer.Flush()
	klog.Infof("IP addresses list %v written to temporary file:%s", ipAddressList, tmpFile.Name())

	//nolint:gosec
	cmd := exec.Command("make", "helm-csi-install", "PROJECT="+project, "IP_LIST_FILE="+tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	klog.Info(string(output))
	return nil
}

func deleteDriver(project string) error {
	envVars := []string{
		"PROJECT=" + project,
	}

	//nolint:gosec
	cmd := exec.Command("make", "helm-csi-uninstall")
	cmd.Env = append(os.Environ(), envVars...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}

	klog.Info(string(output))
	return nil
}
