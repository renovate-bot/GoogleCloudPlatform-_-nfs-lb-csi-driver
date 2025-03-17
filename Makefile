# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GKE_NFS_LB_SCRIPTS_DIR=./gke-nfs-lb
ifeq ($(origin GIT_COMMIT), undefined)
    GIT_COMMIT := $(shell git rev-parse HEAD)
endif
$(info GIT_COMMIT is ${GIT_COMMIT})

BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
IMAGENAME ?= nfsplugin
IMAGE_VERSION ?= ${GIT_COMMIT}
REGISTRY ?= gcr.io
PROJECT ?= $(shell gcloud config get-value project 2>&1 | head -n 1)
NFS_CSI_IMAGE = $(REGISTRY)/$(PROJECT)/$(IMAGENAME)
$(info PROJECT is ${PROJECT})

ARCH = amd64
PKG = github.com/GoogleCloudPlatform/nfs-lb-csi-driver
GINKGO_FLAGS = -ginkgo.v
GO111MODULE = on
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin
DOCKER_CLI_EXPERIMENTAL = enabled
export GOPATH GOBIN GO111MODULE DOCKER_CLI_EXPERIMENTAL

LDFLAGS = -X ${PKG}/pkg/nfs.driverVersion=${IMAGE_VERSION} -X ${PKG}/pkg/nfs.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/nfs.buildDate=${BUILD_DATE}
EXT_LDFLAGS = -s -w -extldflags "-static"

# Helm chart variables
HELM_CHART_VERSION ?= v0.0.1
IP_LIST_FILE ?= "$(GKE_NFS_LB_SCRIPTS_DIR)/ips.txt"
IP_LIST = $(shell $(GKE_NFS_LB_SCRIPTS_DIR)/generate_ip_list.sh $(IP_LIST_FILE))
CSI_NAMESPACE ?= "gke-csi-nfs-lb"
HELM_OPTIONS = $(GKE_NFS_LB_SCRIPTS_DIR)/charts/$(HELM_CHART_VERSION)/nfs-csi-lb --set image.nfs.repository="$(REGISTRY)/$(PROJECT)/$(IMAGENAME)" --set image.nfs.tag=$(IMAGE_VERSION) --set controller.ipaddressList="${IP_LIST}"

.EXPORT_ALL_VARIABLES:

all: nfs

.PHONY: verify
verify: unit-test
	hack/verify-all.sh

.PHONY: unit-test
unit-test:
	go test -covermode=count -coverprofile=profile.cov ./pkg/... -v

.PHONY: sanity-test
sanity-test: nfs
	./test/sanity/run-test.sh

.PHONY: local-build-push
local-build-push: nfs
	docker build -t $(LOCAL_USER)/nfsplugin:latest .
	docker push $(LOCAL_USER)/nfsplugin

.PHONY: nfs
nfs:
	@echo "running nfs..."
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -a -ldflags "${LDFLAGS} ${EXT_LDFLAGS}" -mod vendor -o bin/${ARCH}/nfsplugin ./cmd/nfsplugin

build-nfs-csi-image-and-push: nfs init-buildx
		@echo "running build-nfs-csi-image-and-push..."
		{                                                                   \
		set -e ;                                                            \
		docker buildx build \
			--platform linux/amd64 \
			--build-arg ARCH=$(ARCH) \
			-f ./Dockerfile \
			-t $(NFS_CSI_IMAGE):$(IMAGE_VERSION) --push .; \
		}

init-buildx:
	@echo "running init-buildx..."
	# Ensure we use a builder that can leverage it (the default on linux will not)
	-docker buildx rm multiarch-multiplatform-builder
	docker buildx create --use --name=multiarch-multiplatform-builder
	docker run --rm --privileged multiarch/qemu-user-static --reset --credential yes --persistent yes
	# Register gcloud as a Docker credential helper.
	# Required for "docker buildx build --push".
	gcloud auth configure-docker --quiet

.PHONY: e2e-test
e2e-test:
	if [ ! -z "$(EXTERNAL_E2E_TEST)" ]; then \
		bash ./test/external-e2e/run.sh;\
	else \
		go test -v -timeout=0 ./test/e2e ${GINKGO_FLAGS};\
	fi


.PHONY: e2e-lb-test
e2e-lb-test:
	./test/e2e-lb/run-e2e-local.sh

.PHONY: helm-csi-install
helm-csi-install:
	helm upgrade -i nfs-csi-lb --create-namespace --namespace ${CSI_NAMESPACE} --wait --timeout=15m -v=5 --debug ${HELM_OPTIONS}

.PHONY: helm-csi-uninstall
helm-csi-uninstall:
	helm uninstall nfs-csi-lb --namespace ${CSI_NAMESPACE} --wait
	@echo "Cleanup namespace ${CSI_NAMESPACE}"
	kubectl delete namespace ${CSI_NAMESPACE}
	@echo "Cleanup node assigned-ip annotations"
	$(GKE_NFS_LB_SCRIPTS_DIR)/lb_delete_assigned_ip_node_ann.sh
