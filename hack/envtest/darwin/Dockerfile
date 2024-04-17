#  Copyright 2018 The Kubernetes Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

ARG GO_VERSION
FROM golang:${GO_VERSION} as builder

# Version and platform args.
ARG KUBERNETES_VERSION
ARG ETCD_VERSION
ARG OS=darwin
ARG ARCH

# Tools path.
ENV DEST=/controller-tools/envtest

# Install dependencies.
RUN apt update && \
    apt install unzip rsync -y && \
    mkdir -p $DEST

# kube-apiserver
WORKDIR /kubernetes
RUN git clone https://github.com/kubernetes/kubernetes . --depth=1 -b ${KUBERNETES_VERSION}
ENV CGO_ENABLED=0
ENV KUBE_BUILD_PLATFORMS=${OS}/${ARCH}
RUN make WHAT=cmd/kube-apiserver && \
    cp _output/local/bin/${KUBE_BUILD_PLATFORMS}/kube-apiserver $DEST

# kubectl
RUN /bin/bash -x -c ' \
    { curl -sfLO https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/${OS}/${ARCH}/kubectl && \
      chmod +x kubectl && \
      cp kubectl $DEST; } || \
    { make WHAT=cmd/kubectl && \
      cp _output/local/bin/${KUBE_BUILD_PLATFORMS}/kubectl $DEST; }'

# etcd
ENV ETCD_BASE_NAME=etcd-${ETCD_VERSION}-${OS}-${ARCH}
RUN curl -sfLO https://github.com/coreos/etcd/releases/download/${ETCD_VERSION}/${ETCD_BASE_NAME}.zip && \
    unzip -o ${ETCD_BASE_NAME}.zip && \
    cp ${ETCD_BASE_NAME}/etcd $DEST

# Package into tarball.
RUN tar -czvf /envtest-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz $DEST
RUN sha512sum /envtest-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz > /envtest-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz.sha512

# Build the final image with the binaries.
FROM scratch
ARG OS
ARG ARCH
ARG KUBERNETES_VERSION
COPY --from=builder /envtest-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz /
COPY --from=builder /envtest-${KUBERNETES_VERSION}-${OS}-${ARCH}.tar.gz.sha512 /
