**Note**
The following set of steps has been validated only on a [GKE Standard cluster](https://cloud.google.com/kubernetes-engine/docs/resources/autopilot-standard-feature-comparison)

# Steps to install NFS CSI LB driver via helm

A set of helper commands are provided via ./Makefile rules to simply workflow

## Install Helm

```
curl -fsSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

## Buld the CSI driver container images

```console
PROJECT=<your-gcp-project> IMAGE_VERSION=latest make build-nfs-csi-image-and-push
```
The above command builds the container images and  pushes the container registry

## Install CSI driver

```console
PROJECT=<your-gcp-project> IMAGE_VERSION=latest make helm-csi-install
```

The above command does the following key steps:
1. Creats a k8s namespace based on $CSI_NAMESPACE variable (default `gke-csi-nfs-lb`)
2. Fetch the list of nfs server IPs from a file $IP_LIST_FILE. The default value of this variable is set to ./gke-nfs-lb/ips.txt
3. Prepare helm options to override the controller and node server container images, and generate the comma separated IP list for the controller driver

A fully deployed NFS CSI LB driver shows up as follows (the example is based on a 3 node GKE Cluster):
```
$ kubectl get all -n gke-csi-nfs-lb
NAME                                         READY   STATUS    RESTARTS   AGE
pod/csi-nfs-lb-controller-6b986dfcdf-mfz8t   2/2     Running   0          13s
pod/csi-nfs-lb-node-6w8sh                    2/2     Running   0          13s
pod/csi-nfs-lb-node-l24x7                    2/2     Running   0          13s
pod/csi-nfs-lb-node-v5hs4                    2/2     Running   0          13s

NAME                             DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
daemonset.apps/csi-nfs-lb-node   3         3         3       3            3           kubernetes.io/os=linux   13s

NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/csi-nfs-lb-controller   1/1     1            1           13s

NAME                                               DESIRED   CURRENT   READY   AGE
replicaset.apps/csi-nfs-lb-controller-6b986dfcdf   1         1         1       13s
```

## Uninstall CSI driver

This cleanups the CSI driver, the associated k8s namespace and removes the annotations from the nodes.
```console
make helm-csi-uninstall
```

## CSI Driver update workflow

Commons scenarios include the following

### Update the CSI driver container images

Simply run the helm install with the updated image versions
```console
PROJECT=<your-gcp-project> IMAGE_VERSION=<updated-image-version> make helm-csi-install
```

### Update the NFS Server IP list

**WARNING**
This is a disruptive operation that can impact existing k8s pods. So it is recommended to tear down existing workloads before updating the IP list of the CSI driver

Steps:
1. Cleanup any k8s pods and volumes which is using the CSI driver for mounting the target volumes
2. Uninstall the existing CSI driver and cleanup the node annotations
```console
PROJECT=<your-gcp-project> make helm-csi-install
```
3. Point $IP_LIST_FILE to the updated list of IP(s)
4. Redeploy the CSI driver via helm install
```
PROJECT=<your-gcp-project> IMAGE_VERSION=latest IP_LIST_FILE=</path/to/IP-list-file> make helm-csi-install
```

