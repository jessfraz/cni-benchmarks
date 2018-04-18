FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/jessfraz/cni-benchmarks

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/jessfraz/cni-benchmarks \
	&& make static \
	&& mv cni-benchmarks /usr/bin/cni-benchmarks \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/cni-benchmarks /usr/bin/cni-benchmarks
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "cni-benchmarks" ]
CMD [ "--help" ]
