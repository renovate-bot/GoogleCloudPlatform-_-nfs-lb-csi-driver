# NFS Client Server example

In this example we will deploy a pool of NFS Servers (2 pods), and 3 NFS Clients (a deployment with 3 replicas, one pod per node) and demonstrate how NFS LB CSI driver balances the target NFS Server IP allocated to each of the nfs client pods. The example uses yaml spec from [here](../gke-nfs-lb/examples/nfs-client-server/)

## Deploy NFS Server pool

```
$ NFS_SERVER_NAME=nfs-server-0 envsubst < ./gke-nfs-lb/examples/nfs-client-server/nfs-server-v3.yaml | kubectl apply -f -

$ NFS_SERVER_NAME=nfs-server-1 envsubst < ./gke-nfs-lb/examples/nfs-client-server/nfs-server-v3.yaml | kubectl apply -f -
```

```
$ kubectl get po | grep nfs-server
nfs-server-0-556fbdb58-8r74l    1/1     Running   0          2m
nfs-server-1-676f79dc97-skvdd   1/1     Running   0          110s
```

```
$ kubectl get service | grep nfs-server
nfs-server-0   ClusterIP   34.118.238.111   <none>        2049/TCP,111/TCP,20048/TCP,4045/TCP   21m
nfs-server-1   ClusterIP   34.118.227.183   <none>        2049/TCP,111/TCP,20048/TCP,4045/TCP   21m
```

### Extract the service IP and prepare the IP list file

```
$ kubectl get services nfs-server-0 -o jsonpath='{.spec.clusterIP}'
34.118.238.111

$ kubectl get services nfs-server-1 -o jsonpath='{.spec.clusterIP}'
34.118.227.183
```

```
echo -e "34.118.227.183\n34.118.238.111" > /tmp/ips.txt
```

### Install the CSI driver

Build the CSI container images
```
$ PROJECT=<your-gcp-project>  make build-nfs-csi-image-and-push
```

Deploy the CSI driver
```
$ PROJECT=<your-gcp-project> IP_LIST_FILE=/tmp/ips.txt make helm-csi-install
```

Verify the CSI installation
```
$ kubectl get all -n gke-csi-nfs-lb
NAME                                         READY   STATUS    RESTARTS   AGE
pod/csi-nfs-lb-controller-7854ff45c7-5t9qb   2/2     Running   0          70s
pod/csi-nfs-lb-node-65glb                    2/2     Running   0          70s
pod/csi-nfs-lb-node-w7zl6                    2/2     Running   0          70s
pod/csi-nfs-lb-node-xjrbj                    2/2     Running   0          70s

NAME                             DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
daemonset.apps/csi-nfs-lb-node   3         3         3       3            3           kubernetes.io/os=linux   70s

NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/csi-nfs-lb-controller   1/1     1            1           70s

NAME                                               DESIRED   CURRENT   READY   AGE
replicaset.apps/csi-nfs-lb-controller-7854ff45c7   1         1         1       70s
```

### Deploy the NFS Client pods

```
$ kubectl apply -f ./gke-nfs-lb/examples/nfs-client-server/nfs-client-v3.yaml
```
```
$ kubectl get po -o wide | grep ubuntu
ubuntu-7566fc45cd-b4mv4         1/1     Running   0          37s     10.124.0.18   gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-4fwl   <none>           <none>
ubuntu-7566fc45cd-j29cg         1/1     Running   0          37s     10.124.1.4    gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-7v8f   <none>           <none>
ubuntu-7566fc45cd-jjqxd         1/1     Running   0          37s     10.124.2.7    gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-33t9   <none>           <none>
```

### Check IP distribution

Verify from the CSI driver logs the IP used for the mount operation

```
$ kubectl logs csi-nfs-lb-node-65glb -n gke-csi-nfs-lb -c nfs
I0721 01:58:38.619907       1 mount_linux.go:224] Mounting cmd (mount) with arguments (-t nfs -o vers=3 34.118.227.183:/exports /var/lib/kubelet/pods/ad74d245-3921-4882-890d-6ad0f0ac987c/volumes/kubernetes.io~csi/nfs-client-pv/mount)
I0721 01:58:41.753394       1 nodeserver.go:168] skip chmod on targetPath(/var/lib/kubelet/pods/ad74d245-3921-4882-890d-6ad0f0ac987c/volumes/kubernetes.io~csi/nfs-client-pv/mount) since mountPermissions is set as 0
I0721 01:58:41.753423       1 nodeserver.go:182] volume(nfs-server.default.svc.cluster.local/share##) mount 34.118.227.183:/exports on /var/lib/kubelet/pods/ad74d245-3921-4882-890d-6ad0f0ac987c/volumes/kubernetes.io~csi/nfs-client-pv/mount succeeded
I0721 01:58:41.753436       1 utils.go:123] GRPC response: {}
```

```
$ kubectl logs csi-nfs-lb-node-w7zl6 -n gke-csi-nfs-lb -c nfs
I0721 01:58:48.732841       1 mount_linux.go:224] Mounting cmd (mount) with arguments (-t nfs -o vers=3 34.118.227.183:/exports /var/lib/kubelet/pods/0ee2100d-d6ee-4f02-a2ea-af6538346c6c/volumes/kubernetes.io~csi/nfs-client-pv/mount)
I0721 01:58:51.746808       1 nodeserver.go:168] skip chmod on targetPath(/var/lib/kubelet/pods/0ee2100d-d6ee-4f02-a2ea-af6538346c6c/volumes/kubernetes.io~csi/nfs-client-pv/mount) since mountPermissions is set as 0
I0721 01:58:51.746845       1 nodeserver.go:182] volume(nfs-server.default.svc.cluster.local/share##) mount 34.118.227.183:/exports on /var/lib/kubelet/pods/0ee2100d-d6ee-4f02-a2ea-af6538346c6c/volumes/kubernetes.io~csi/nfs-client-pv/mount succeeded
I0721 01:58:51.746860       1 utils.go:123] GRPC response: {}
```

```
$ kubectl logs csi-nfs-lb-node-xjrbj -n gke-csi-nfs-lb -c nfs
I0721 01:58:36.341966       1 mount_linux.go:224] Mounting cmd (mount) with arguments (-t nfs -o vers=3 34.118.238.111:/exports /var/lib/kubelet/pods/4d79461b-25a8-4f4d-a2ac-e3dabf10c8f0/volumes/kubernetes.io~csi/nfs-client-pv/mount)
I0721 01:58:39.649743       1 nodeserver.go:168] skip chmod on targetPath(/var/lib/kubelet/pods/4d79461b-25a8-4f4d-a2ac-e3dabf10c8f0/volumes/kubernetes.io~csi/nfs-client-pv/mount) since mountPermissions is set as 0
I0721 01:58:39.649776       1 nodeserver.go:182] volume(nfs-server.default.svc.cluster.local/share##) mount 34.118.238.111:/exports on /var/lib/kubelet/pods/4d79461b-25a8-4f4d-a2ac-e3dabf10c8f0/volumes/kubernetes.io~csi/nfs-client-pv/mount succeeded
I0721 01:58:39.649793       1 utils.go:123] GRPC response: {}
```


Check node annotations. We can see `34.118.227.183` is used for 2 of the pods, and `34.118.238.111` used for 1 pod
```
$ kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}:{.metadata.annotations.nfs\.lb\.csi\.storage\.gke\.io/assigned-ip}{"\n"}{end}'
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-33t9:34.118.227.183
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-4fwl:34.118.227.183
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-7v8f:34.118.238.111
```

### Tear down the NFS Client pods

```
$ kubectl delete -f ./gke-nfs-lb/examples/nfs-client-server/nfs-client-v3.yaml
persistentvolume "nfs-client-pv" deleted
persistentvolumeclaim "nfs-client-pvc" deleted
deployment.apps "ubuntu" deleted
```

Verify node annotation is cleared
```
$ kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}:{.metadata.annotations.nfs\.lb\.csi\.storage\.gke\.io/assigned-ip}{"\n"}{end}'
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-33t9:
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-4fwl:
gke-cluster-nfs-csi-e2e--default-pool-c71b8b02-7v8f:
```

### Teadown NFS Servers

```
$ NFS_SERVER_NAME=nfs-server-0 envsubst < ./gke-nfs-lb/examples/nfs-client-server/nfs-server-v3.yaml | kubectl delete -f -
$ NFS_SERVER_NAME=nfs-server-1 envsubst < ./gke-nfs-lb/examples/nfs-client-server/nfs-server-v3.yaml | kubectl delete -f -
```

### Uninstall CSI driver

```
PROJECT=<your-gcp-project> make helm-csi-uninstall
```