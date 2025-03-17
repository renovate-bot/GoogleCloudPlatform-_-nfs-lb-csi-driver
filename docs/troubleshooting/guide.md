# Troubleshooting

# Overview of the CSI driver
## Controller and node driver components
The driver is deployed via helm installation command `helm-csi-install` in [Makefile](../../Makefile). This would deploy the driver to the k8s namespace controller by the $CSI_NAMESPACE variable. The default value is `gke-csi-nfs-lb`

To list all the driver pods, run
```console
$ kubectl get po -n gke-csi-nfs-lb -o wide
NAME                                     READY   STATUS    RESTARTS   AGE   IP              NODE                                             NOMINATED NODE   READINESS GATES
csi-nfs-lb-controller-5c9cb4c6c7-np7m7   2/2     Running   0          22m   10.12.46.6      gke-cluster-nfs-csi-default-pool-957a01d7-jbbc   <none>           <none>
csi-nfs-lb-node-hrdx9                    2/2     Running   0          40m   10.128.15.194   gke-cluster-nfs-csi-default-pool-957a01d7-6pbk   <none>           <none>
csi-nfs-lb-node-jtp48                    2/2     Running   0          40m   10.128.15.220   gke-cluster-nfs-csi-default-pool-957a01d7-jbbc   <none>           <none>
csi-nfs-lb-node-kp7w2                    2/2     Running   0          40m   10.128.15.213   gke-cluster-nfs-csi-default-pool-957a01d7-434j   <none>           <none>
csi-nfs-lb-node-sq9d5                    2/2     Running   0          83m   10.128.0.20     gke-cluster-nfs-csi-default-pool-957a01d7-i0a0   <none>           <none>
csi-nfs-lb-node-zmwdm                    2/2     Running   0          39m   10.128.15.208   gke-cluster-nfs-csi-default-pool-957a01d7-xgxp   <none>           <none>
```
The above example shows 1 controller driver pod `csi-nfs-lb-controller-5c9cb4c6c7-np7m7` and 5 node driver pods `csi-nfs-lb-node-*` for a 5 node GKE cluster

## Driver logs
### Query Controller driver pod logs
Controller driver contains 2 containers: csi-attacher and nfs. The container logs can be queried as follows
1. csi-attacher
```
 kubectl logs csi-nfs-lb-controller-5c9cb4c6c7-np7m7  -c csi-attacher -n gke-csi-nfs-lb
```

2. nfs
```
 kubectl logs csi-nfs-lb-controller-5c9cb4c6c7-np7m7  -c nfs -n gke-csi-nfs-lb
```
### Query Node driver pod logs

1. nfs container logs
```
$ kubectl logs csi-nfs-lb-node-hrdx9   -c nfs -n gke-csi-nfs-lb
```

### Veirfy the commit ID used by the containers
The controller and node driver logs for the `nfs` container spits out the git commit ID used.
```
Build Date: "2024-07-17T17:44:54Z"
Compiler: gc
Driver Name: nfs.lb.csi.storage.gke.io
Driver Version: 03509d0f9c231f9e9a0ea34181771925fdecb754
Git Commit: 03509d0f9c231f9e9a0ea34181771925fdecb754
Go Version: go1.23-20240626-RC01 cl/646990413 +5a18e79687 X:fieldtrack,boringcrypto
Platform: linux/amd64
```

### Verify NFS services started

If pods fail to mount, and stuck in container creation, check if the node driver has successfully started the nfs services daemon. The node driver container `nfs` invokes the [nfs_services_start.sh](../../gke-nfs-lb/nfs_services_start.sh) to start the necessary nfs services and daemons. A successful script launch looks as follows
```
$ kubectl logs csi-nfs-lb-node-2d4gd -c nfs -n gke-csi-nfs-lb 
...
Starting RPC port mapper daemon: rpcbind.
program 100024 version 1 ready and waiting
statd already running
Starting NFS common utilities: idmapd.
...
```

### Check IP map update during ControllerPublish

The CSI driver maintains an in-memory map of IP to node counts. On every CSI ControllerPublishVolume call, the keys of the map are sorted by node count, and the smallest count IP key is chosen as the target IP for the given volume on that given node.  The IP is also stamped on the given node object with an annotation `nfs.lb.csi.storage.gke.io/assigned-ip`. The logs from the controller driver pod's `nfs` container can be seen as follows. It shows a snippet where for given node `gke-cluster-nfs-csi-default-pool-957a01d7-xgxp` and volumeID `nfs-server.default.svc.cluster.local/vol1`, IP `10.94.112.74` was chosen and updated. In the IPMap the value `3` indicates, 3 GKE nodes have been alloted the IP

```
I0717 19:41:17.901569       1 utils.go:116] GRPC call: /csi.v1.Controller/ControllerPublishVolume
I0717 19:41:17.901661       1 utils.go:117] GRPC request: {"node_id":"gke-cluster-nfs-csi-default-pool-957a01d7-434j","volume_capability":{"AccessType":{"Mount":{"mount_flags":["vers=3"]}},"access_mode":{"mode":5}},"volume_context":{"share":"/vol1"},"volume_id":"nfs-server.default.svc.cluster.local/vol1"}
I0717 19:41:17.923377       1 round_trippers.go:553] PUT https://34.118.224.1:443/api/v1/nodes/gke-cluster-nfs-csi-default-pool-957a01d7-xgxp 200 OK in 23 milliseconds
I0717 19:41:17.928433       1 lbcontroller.go:151] AssignIPToNode: For volume "nfs-server.default.svc.cluster.local/vol1", node "gke-cluster-nfs-csi-default-pool-957a01d7-xgxp", IP updated "10.94.112.74", LB controller IP map map[10.94.112.74:2]
I0717 19:41:17.928648       1 utils.go:123] GRPC response: {"publish_context":{"nfs.lb.csi.storage.gke.io/assigned-ip":"10.94.112.74"}}
I0717 19:41:17.929113       1 lbcontroller.go:143] Assigning IP "10.94.112.74" to node "gke-cluster-nfs-csi-default-pool-957a01d7-434j" for volume "nfs-server.default.svc.cluster.local/vol1"
I0717 19:41:17.948177       1 round_trippers.go:553] PUT https://34.118.224.1:443/api/v1/nodes/gke-cluster-nfs-csi-default-pool-957a01d7-434j 200 OK in 17 milliseconds
I0717 19:41:17.952504       1 lbcontroller.go:151] AssignIPToNode: For volume "nfs-server.default.svc.cluster.local/vol1", node "gke-cluster-nfs-csi-default-pool-957a01d7-434j", IP updated "10.94.112.74", LB controller IP map map[10.94.112.74:3]
I0717 19:41:17.952553       1 utils.go:123] GRPC response: {"publish_context":{"nfs.lb.csi.storage.gke.io/assigned-ip":"10.94.112.74"}}
```

#### Verify IP load balancing
While the pods are actively mounted, we can verify the annotations of the node by running the helper script [lb_list_assigned_ip_node_ann.sh](../../gke-nfs-lb/lb_list_assigned_ip_node_ann.sh)

Consider a scenario where 3 pods are mounted on 3 out of 5 GKE nodes

```console
$ kubectl get po -o wide
NAME                      READY   STATUS    RESTARTS   AGE     IP            NODE                                             NOMINATED NODE   READINESS GATES
ubuntu-7fb5c94594-lvwkg   1/1     Running   0          9m29s   10.12.46.9    gke-cluster-nfs-csi-default-pool-957a01d7-jbbc   <none>           <none>
ubuntu-7fb5c94594-w52g8   1/1     Running   0          9m29s   10.12.38.7    gke-cluster-nfs-csi-default-pool-957a01d7-434j   <none>           <none>
ubuntu-7fb5c94594-z8hlh   1/1     Running   0          9m29s   10.12.33.10   gke-cluster-nfs-csi-default-pool-957a01d7-xgxp   <none>           <none>
```

The node annotation would be seen `nfs.lb.csi.storage.gke.io/assigned-ip` would be seen only on the corresponding nodes where the pods are mounted.
```console

csi-driver-nfs-lb/gke-nfs-lb$ ./lb_list_assigned_ip_node_ann.sh 
gke-cluster-nfs-csi-default-pool-957a01d7-434j 10.94.112.74
gke-cluster-nfs-csi-default-pool-957a01d7-6pbk 
gke-cluster-nfs-csi-default-pool-957a01d7-i0a0 
gke-cluster-nfs-csi-default-pool-957a01d7-jbbc 10.94.112.74
gke-cluster-nfs-csi-default-pool-957a01d7-xgxp 10.94.112.74

```

We can also run another helper script to understand the distribution of IP to nodes. The output shows that `10.94.112.74` is assigned to 3 GKE nodes

```console
csi-driver-nfs-lb/gke-nfs-lb$ ./lb_group_nodes_per_ip.sh 
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-6pbk
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-i0a0
IP: 10.94.112.74, Count: 3
```

### Check IP map update during ControllerUnublishVolume
During unmount, the controllerUnpublishVolume CSI call is invoked which removes the annotation from the nodes
```
I0717 19:56:02.352435       1 utils.go:116] GRPC call: /csi.v1.Controller/ControllerUnpublishVolume
I0717 19:56:02.352665       1 utils.go:117] GRPC request: {"node_id":"gke-cluster-nfs-csi-default-pool-957a01d7-xgxp","volume_id":"nfs-server.default.svc.cluster.local/vol1"}
I0717 19:56:02.352939       1 lbcontroller.go:179] Removing IP annotation "10.94.112.74" from node "gke-cluster-nfs-csi-default-pool-957a01d7-xgxp" for volume "nfs-server.default.svc.cluster.local/vol1"
I0717 19:56:02.371299       1 round_trippers.go:553] PUT https://34.118.224.1:443/api/v1/nodes/gke-cluster-nfs-csi-default-pool-957a01d7-xgxp 200 OK in 17 milliseconds
I0717 19:56:02.373236       1 lbcontroller.go:188] RemoveIPFromNode: For volume "nfs-server.default.svc.cluster.local/vol1", node "gke-cluster-nfs-csi-default-pool-957a01d7-xgxp", IP updated "10.94.112.74", LB controller IP map map[10.94.112.74:0]
I0717 19:56:02.373467       1 controllerserver.go:330] ControllerUnpublishVolume succeeded for volume nfs-server.default.svc.cluster.local/vol1 from node gke-cluster-nfs-csi-default-pool-957a01d7-xgxp
I0717 19:56:02.373600       1 utils.go:123] GRPC response: {}
```

Now running the script again, shows no annotation detected on the GKE nodes, since all pods were unmounted.

```
csi-driver-nfs-lb/gke-nfs-lb$ ./lb_group_nodes_per_ip.sh 
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-434j
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-6pbk
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-i0a0
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-jbbc
Skipping line: No valid IP detected - gke-cluster-nfs-csi-default-pool-957a01d7-xgxp
```

### Debug controller driver pod crashloop

If the controller driver pod reports a crashloop, check logs. Controller may crash if no valid IP list is detected

```
$ kubectl logs csi-nfs-lb-controller-7dbf6847c9-8svlz -c nfs -n gke-csi-nfs-lb
I0718 20:15:22.946840       1 main.go:60] runController true, runNodeServer false, runNfsServices false
F0718 20:15:22.947194       1 main.go:106] NFS LB CSI Driver controller requires atleast one valid IP address for the target NFS server(s)
I0718 20:15:22.947707       1 main.go:83] waiting for SIGTERM signal...
```

A quick check of the controller pod spec shows that the IP list is empty.
```
$ kubectl get po csi-nfs-lb-controller-7dbf6847c9-8svlz -n gke-csi-nfs-lb -o yaml | grep ip-address
    - --ip-addresses=
```

Check if the file provided in $IP_LIST_FILE is setup correct, that is used during the [CSI driver installation](../helm-install-csi-driver.md) step
```
PROJECT=<your-gcp-project> IP_LIST_FILE=<ip-list-file>  make helm-csi-install
```