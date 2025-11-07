#!/usr/bin/env bash

set -x
set -e


kind delete cluster
kind create cluster

kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-crds.yaml

kubectl wait --for=condition=Established --all crd -l knative.dev/crd-install="true"

kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-core.yaml
kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/latest/download/release.yaml
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/in-memory-channel.yaml
kubectl apply -f https://github.com/knative/eventing/releases/latest/download/mt-channel-broker.yaml
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.yaml

kubectl patch svc broker-ingress -n knative-eventing -p '{"spec": {"type": "LoadBalancer"}}'

kubectl wait --for=condition=Available deployment -A --all --timeout 2m

kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

kubectl patch configmap/config-domain \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"margarita.dev":""}}'

cat <<EOF | kubectl apply -f -
# this issuer is used by cert-manager to sign all certificates
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cluster-selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer # this issuer is specifically for Knative, it will use the CA stored in the secret created by the Certificate below
metadata:
  name: knative-selfsigned-issuer
spec:
  ca:
    secretName: knative-selfsigned-ca
---
apiVersion: cert-manager.io/v1
kind: Certificate # this creates a CA certificate, signed by cluster-selfsigned-issuer and stored in the secret knative-selfsigned-ca
metadata:
  name: knative-selfsigned-ca
  namespace: cert-manager #  If you want to use it as a ClusterIssuer the secret must be in the cert-manager namespace.
spec:
  secretName: knative-selfsigned-ca
  commonName: knative.dev
  usages:
    - server auth
  isCA: true
  issuerRef:
    kind: ClusterIssuer
    name: cluster-selfsigned-issuer
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-certmanager
  namespace: knative-serving
  labels:
    networking.knative.dev/certificate-provider: cert-manager
data:
  issuerRef: |
    kind: ClusterIssuer
    name: knative-selfsigned-issuer
  clusterLocalIssuerRef: |
    kind: ClusterIssuer
    name: knative-selfsigned-issuer
  systemInternalIssuerRef: |
    kind: ClusterIssuer
    name: knative-selfsigned-issuer
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-network
  namespace: knative-serving
data:
  external-domain-tls: Enabled
---
apiVersion: v1
kind: ConfigMap
metadata:
 name: config-autoscaler
 namespace: knative-serving
data:
 stable-window: "6s"
EOF

kubectl rollout restart deploy/controller -n knative-serving
