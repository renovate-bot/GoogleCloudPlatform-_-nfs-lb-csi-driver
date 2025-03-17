
Steps to run the e2e tests for NFS LB CSI Driver
## Create a GKE cluster 
```
gcloud container --project <your-gcp-project> clusters create "cluster-nfs-csi-e2e" --zone "us-central1-c"  --release-channel "regular" --machine-type "n2-standard-16"  --num-nodes "10"
```

## Setup kubeconfig for the cluster
```
gcloud container clusters get-credentials  cluster-nfs-csi-e2e  --location=us-central1-c
```

## Start the e2e test
The e2e test performs the following steps
1. Deploy a pool of NFS server pods in the cluster
2. Build and deploy the NFS LB CSI driver with the IP list of the nfs server pool
3. Spin up test-cases
4. Teardown the CSI driver and NFS server pools after the test completes

### Test case 1
10 GKE nodes, 10 nfs server, 10 nfs client

```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=10 E2E_TEST_FOCUS="bringup-teardown-10node-10nfsclient-10nfsserver" make e2e-lb-test
```

### Test case 2
10 GKE nodes, 5 nfs server, 5 nfs client
```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=5 E2E_TEST_FOCUS="bringup-teardown-10node-5nfsclient-5nfsserver" make e2e-lb-test
```

### Test case 3
10 GKE nodes, 5 nfs server, 10 nfs client
```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=5 E2E_TEST_FOCUS="bringup-teardown-10node-10nfsclient-5nfsserver" make e2e-lb-test
```

### Test case 4
10 GKE nodes, 5 nfs server, 9 nfs client
```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=5 E2E_TEST_FOCUS="bringup-teardown-10node-9nfsclient-5nfsserver" make e2e-lb-test
```

### Test case 5
10 GKE nodes, 5 nfs server, 2 5replica nfs client
```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=5 E2E_TEST_FOCUS="bringup-teardown-10node-2-5nfsclient-5nfsserver" make e2e-lb-test
```


### Test case 6

100 GKE node, 10 nfs server, 100 nfs client

```
PROJECT=<your-gcp-project> E2E_INSTALL_CSI_DRIVER=true E2E_DESTROY_CSI_DRIVER=true E2E_BUILD_CSI_DRIVER_IMAGE=true E2E_TEST_NFS_SERVER_COUNT=10 E2E_TEST_FOCUS="bringup-teardown-100node-100nfsclient-10nfsserver" make e2e-lb-test
```