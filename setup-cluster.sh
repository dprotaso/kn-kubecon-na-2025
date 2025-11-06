#!/usr/bin/env bash

set -x
set -e


kind delete cluster
kind create cluster

# Serving
kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml
kubectl wait --for=condition=Established --all crd -l knative.dev/crd-install="true"

kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml
kubectl wait --for=condition=Available deployment -n knative-serving --all --timeout 1m

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"margarita.dev":""}}'

kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/latest/download/release.yaml
kubectl wait --for=condition=Available deployment -n knative-serving --all --timeout 1m

# Eventing
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-crds.yaml
kubectl wait --for=condition=Established --all crd -l knative.dev/crd-install="true"

kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-core.yaml
kubectl wait --for=condition=Available deployment -n knative-eventing --all --timeout 1m

kubectl apply -f https://github.com/knative/eventing/releases/latest/download/in-memory-channel.yaml
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/mt-channel-broker.yaml
kubectl patch svc broker-ingress -n knative-eventing -p '{"spec": {"type": "LoadBalancer"}}'
kubectl wait --for=condition=Available deployment -n knative-eventing --all --timeout 1m
