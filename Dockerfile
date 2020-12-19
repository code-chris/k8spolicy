FROM golang:1.15.6 as builder

WORKDIR /go/src/app
COPY . .

RUN go build -ldflags '-w -s' -o /k8spolicy


FROM debian:buster-slim

ENV CONFTEST_VERSION 0.20.0
ENV K8SPOLICY_SKIP_POLICY_DOWNLOAD true
ENV K8SPOLICY_SKIP_CONFTEST_DOWNLOAD true

RUN apt-get update && \
    apt-get install -y wget ca-certificates --no-install-recommends && \

    mkdir -p /download /tmp/k8spolicy/policies/k8s-api-deprecation && \
    wget https://github.com/swade1987/deprek8ion/archive/master.tar.gz -O /download/deprek8ion.tar.gz && \
    tar xzf /download/deprek8ion.tar.gz -C /download && \
    cp /download/deprek8ion-master/policies/*.rego /tmp/k8spolicy/policies/k8s-api-deprecation && \
    rm -rf /download && \

    mkdir -p /download /tmp/k8spolicy/policies/k8s-security/lib && \
    wget https://github.com/instrumenta/policies/archive/master.tar.gz -O /download/policies.tar.gz && \
    tar xzf /download/policies.tar.gz -C /download && \
    cp /download/policies-master/kubernetes/*.rego /tmp/k8spolicy/policies/k8s-security && \
    cp /download/policies-master/kubernetes/lib/*.rego /tmp/k8spolicy/policies/k8s-security/lib && \
    rm -rf /download && \

    mkdir -p /download && \
    wget https://github.com/instrumenta/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_linux_x86_64.tar.gz -O /download/conftest.tar.gz && \
    tar xzf /download/conftest.tar.gz -C /download && \
    cp /download/conftest /tmp/k8spolicy && \
    rm -rf /download && \

    apt-get upgrade -y && \
    apt-get remove -y wget && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/* && \

    addgroup --gid 1000 k8spolicy && \
    adduser --uid 1000 --gid 1000 --shell /bin/sh --disabled-password --gecos "" k8spolicy && \
    chown -R 1000:1000 /tmp/k8spolicy

COPY --from=builder /k8spolicy /usr/local/bin/k8spolicy
USER k8spolicy
ENTRYPOINT ["/usr/local/bin/k8spolicy"]
