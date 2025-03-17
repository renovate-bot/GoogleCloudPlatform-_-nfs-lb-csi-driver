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
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	e2eframework "k8s.io/kubernetes/test/e2e/framework"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	storageframework "k8s.io/kubernetes/test/e2e/storage/framework"
)

type NFSLBCSITestDriver struct {
	driverInfo storageframework.DriverInfo
}

func InitNFSLBCSITestDriver(bl string, skipGcpSaTest bool) storageframework.TestDriver {
	return &NFSLBCSITestDriver{
		driverInfo: storageframework.DriverInfo{
			Name:        "NFS LB CSI",
			MaxFileSize: storageframework.FileSizeLarge,
			SupportedFsType: sets.NewString(
				"", // Default fsType
			),
			Capabilities: map[storageframework.Capability]bool{
				storageframework.CapPersistence: true,
				storageframework.CapExec:        true,
			},
		},
	}
}

var (
	_ storageframework.TestDriver = &NFSLBCSITestDriver{}
)

func (n *NFSLBCSITestDriver) GetDriverInfo() *storageframework.DriverInfo {
	return &n.driverInfo
}

func (n *NFSLBCSITestDriver) SkipUnsupportedTest(pattern storageframework.TestPattern) {
	if pattern.VolType == storageframework.InlineVolume || pattern.VolType == storageframework.GenericEphemeralVolume {
		e2eskipper.Skipf("NFS LB CSI Driver does not support %s -- skipping", pattern.VolType)
	}
}

func (n *NFSLBCSITestDriver) PrepareTest(ctx context.Context, f *e2eframework.Framework) *storageframework.PerTestConfig {
	config := &storageframework.PerTestConfig{
		Driver:    n,
		Framework: f,
	}
	return config
}

func (n *NFSLBCSITestDriver) CreateVolume(ctx context.Context, config *storageframework.PerTestConfig, volType storageframework.TestVolType) storageframework.TestVolume {
	return nil
}

func (n *NFSLBCSITestDriver) GetPersistentVolumeSource(readOnly bool, _ string, volume storageframework.TestVolume) (*corev1.PersistentVolumeSource, *corev1.VolumeNodeAffinity) {
	return &corev1.PersistentVolumeSource{}, nil
}

func (n *NFSLBCSITestDriver) GetVolume(config *storageframework.PerTestConfig, _ int) (map[string]string, bool, bool) {
	return map[string]string{}, false, false
}

func (n *NFSLBCSITestDriver) GetCSIDriverName(_ *storageframework.PerTestConfig) string {
	return n.driverInfo.Name
}
