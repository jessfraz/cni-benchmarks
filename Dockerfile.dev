FROM golang:stretch as gobuilder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN go clean -i net \
	&& go install -tags netgo std \
	&& go install -race -tags netgo std

FROM gobuilder AS cnibuilder
ENV CNI_VERSION master
RUN git clone --branch "${CNI_VERSION}" --depth 1 https://github.com/containernetworking/plugins.git /go/src/github.com/containernetworking/plugins
WORKDIR /go/src/github.com/containernetworking/plugins
RUN ./build.sh

FROM gobuilder AS azurebuilder
ENV CNI_VERSION master
RUN git clone --branch "${CNI_VERSION}" --depth 1 https://github.com/Azure/azure-container-networking.git /go/src/github.com/Azure/azure-container-networking
WORKDIR /go/src/github.com/Azure/azure-container-networking
RUN go get -d ./cni/...
RUN CGO_ENABLED=0 go build -v -i -o /usr/bin/azure-vnet \
	-ldflags "-s -w" ./cni/network/plugin
RUN CGO_ENABLED=0 go build -v -i -o /usr/bin/azure-vnet-ipam \
	-ldflags "-s -w" ./cni/ipam/plugin

FROM gobuilder AS calicobuilder
ENV CNI_VERSION master
RUN git clone --branch "${CNI_VERSION}" --depth 1 https://github.com/projectcalico/cni-plugin.git /go/src/github.com/projectcalico/cni-plugin
WORKDIR /go/src/github.com/projectcalico/cni-plugin
RUN go get github.com/Masterminds/glide
RUN glide install -strip-vendor
RUN CGO_ENABLED=0 go build -v -i -o /usr/bin/calico \
	-ldflags "-s -w" calico.go
RUN CGO_ENABLED=0 go build -v -i -o /usr/bin/calico-ipam \
	-ldflags "-s -w" ipam/calico-ipam.go

FROM gobuilder AS ciliumbuilder
ENV CILIUM_VERSION master
RUN git clone --branch "${CILIUM_VERSION}" --depth 1 https://github.com/cilium/cilium.git /go/src/github.com/cilium/cilium
WORKDIR /go/src/github.com/cilium/cilium/plugins/cilium-cni
RUN sed -i -e "s/-ldflags '/-ldflags '-s -w /g" ../../Makefile.defs
RUN make

FROM weaveworks/weave AS weavebuilder

FROM alpine:latest
RUN	apk add --no-cache \
	bash
COPY --from=cnibuilder /go/src/github.com/containernetworking/plugins/bin/ /cni/bin/
COPY --from=azurebuilder /usr/bin/azure-vnet /cni/bin/
COPY --from=azurebuilder /usr/bin/azure-vnet-ipam /cni/bin/
COPY --from=calicobuilder /usr/bin/calico /cni/bin/
COPY --from=calicobuilder /usr/bin/calico-ipam /cni/bin/
COPY --from=ciliumbuilder /go/src/github.com/cilium/cilium/plugins/cilium-cni/cilium-cni /cni/bin/
COPY --from=weavebuilder /usr/bin/weaveutil /cni/bin/
WORKDIR /cni/bin
