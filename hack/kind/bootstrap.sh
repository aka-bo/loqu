#!/bin/bash

create_cluster() {
    # kind create cluster --image kindest/node:v1.11.10@sha256:176845d919899daef63d0dbd1cf62f79902c38b8d2a86e5fa041e491ab795d33 --config 3-workers.yaml
    kind create cluster --image kindest/node:v1.12.10@sha256:e43003c6714cc5a9ba7cf1137df3a3b52ada5c3f2c77f8c94a4d73c82b64f6f3 --config 3-workers.yaml
    # kind create cluster --image kindest/node:v1.13.10@sha256:2f5f882a6d0527a2284d29042f3a6a07402e1699d792d0d5a9b9a48ef155fa2a --config 3-workers.yaml
}

load_images() {
    set -x
    kind load docker-image 795669731331.dkr.ecr.us-east-1.amazonaws.com/appsol/loqu:0.0.1
    set +x
}

helm_tillerless_init() {
    echo "helm tiller stop";
    helm tiller stop;
    echo "export HELM_HOST=127.0.0.1:44134";
    export HELM_HOST=127.0.0.1:44134;
    echo "helm tiller start-ci";
    helm tiller start-ci
}

install_apps() {
    export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"

    echo "installing nginx-ingress"
    kubectl apply -f ingress-nginx.yaml

    echo "installing loqu"
    kubectl create namespace testing
    kubectl -n testing apply -f deploy/
    # kubectl -n test-1 apply -f test-1-ingress.yaml
    # kubectl -n test-1 apply -f test-1-idler.yaml
}

create_cluster
load_images
install_apps
