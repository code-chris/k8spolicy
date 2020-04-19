FROM golang:1.14 as builder

WORKDIR /go/src/app
COPY . .

RUN go build -ldflags '-w -s' -o /k8spolicy


FROM debian:buster-slim
COPY --from=builder /k8spolicy /k8spolicy

ENV CONFTEST_VERSION 0.18.1

RUN apt-get update && \
    apt-get install -y wget ca-certificates --no-install-recommends && \
    mkdir -p /download /tmp/k8spolicy/policies && \
    wget https://github.com/swade1987/deprek8ion/archive/master.tar.gz -O /download/master.tar.gz && \
    tar xzf /download/master.tar.gz -C /download && \
    cp /download/deprek8ion-master/policies/*.rego /tmp/k8spolicy/policies && \
    wget https://github.com/instrumenta/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_linux_x86_64.tar.gz -O /download/conftest.tar.gz && \
    tar xzf /download/conftest.tar.gz -C /download && \
    cp /download/conftest /tmp/k8spolicy && \
    rm -rf /download && \
    apt-get remove -y wget && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/* && \
    addgroup --gid 1000 k8spolicy && \
    adduser --uid 1000 --gid 1000 --shell /bin/sh --disabled-password --gecos "" k8spolicy

USER k8spolicy
ENTRYPOINT ["/k8spolicy"]
