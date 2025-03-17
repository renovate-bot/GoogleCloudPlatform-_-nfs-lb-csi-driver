/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GoogleCloudPlatform/nfs-lb-csi-driver/pkg/nfs"

	"k8s.io/klog/v2"
)

var (
	endpoint                     = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID                       = flag.String("nodeid", "", "node id")
	mountPermissions             = flag.Uint64("mount-permissions", 0, "mounted folder permissions")
	driverName                   = flag.String("drivername", nfs.DefaultDriverName, "name of the driver")
	workingMountDir              = flag.String("working-mount-dir", "/tmp", "working directory for provisioner to mount nfs shares temporarily")
	defaultOnDeletePolicy        = flag.String("default-ondelete-policy", "", "default policy for deleting subdirectory when deleting a volume")
	volStatsCacheExpireInMinutes = flag.Int("vol-stats-cache-expire-in-minutes", 10, "The cache expire time in minutes for volume stats cache")
	enableNodeLB                 = flag.Bool("enable-node-lb", false, "When enabled, an external load balancer will assign NFS server IPs to each node. This only works for a single NFS instance")
	ipAddresses                  = flag.String("ip-addresses", "", "Comma-separated list of NFS server IP addresses")
	runControllerServer          = flag.Bool("run-controller-server", false, "if true, starts the controller server")
	runNodeServer                = flag.Bool("run-node-server", false, "if true, starts the node server")
	runNfsServices               = flag.Bool("run-nfs-services", false, "starts NFS services")
)

const (
	NfsServicesStartCmd = "/nfs_services_start.sh"
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	klog.V(4).Infof("runController %v, runNodeServer %v, runNfsServices %v", *runControllerServer, *runNodeServer, *runNfsServices)
	ctx, cancel := context.WithCancel(context.Background())
	if *runNfsServices && !*runControllerServer {
		// Start the NFS services in the background
		cmd := exec.CommandContext(ctx, NfsServicesStartCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Cancel = func() error {
			klog.V(4).Infof("sending SIGTERM to nfs process: %v", cmd)
			return cmd.Process.Signal(syscall.SIGTERM)
		}
		if err := cmd.Start(); err != nil {
			klog.Fatalf("Error starting nfs services: %v", err)
			return
		}

		klog.V(2).Infof("nfs services started in the background with PID: %d", cmd.Process.Pid)
	}

	go handle()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	klog.Info("waiting for SIGTERM signal...")

	<-c // blocking the process
	klog.Info("received SIGTERM signal, calling cancel")
	cancel()

	os.Exit(0)
}

func handle() {
	driverOptions := nfs.DriverOptions{
		NodeID:                       *nodeID,
		DriverName:                   *driverName,
		Endpoint:                     *endpoint,
		MountPermissions:             *mountPermissions,
		WorkingMountDir:              *workingMountDir,
		DefaultOnDeletePolicy:        *defaultOnDeletePolicy,
		VolStatsCacheExpireInMinutes: *volStatsCacheExpireInMinutes,
		RunControllerServer:          *runControllerServer,
		RunNodeServer:                *runNodeServer,
	}

	if *runControllerServer && *ipAddresses == "" {
		klog.Fatal("NFS LB CSI Driver controller requires atleast one valid IP address for the target NFS server(s)")
		return
	}

	ipList := strings.Split(*ipAddresses, ",")
	driverOptions.IPList = ipList
	d := nfs.NewDriver(&driverOptions)
	d.Run(false)
}
