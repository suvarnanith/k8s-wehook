# k8s-webhook

A Kubernetes admission webhook server that validates and mutates `Pod` and
workload resources, written in Go.

## What it does

### Mutating webhook (`/mutate`)

Adds the label `environment=default` to any object (Pod, Deployment,
StatefulSet, DaemonSet, Job, CronJob) that doesn't already declare an
`environment` label.

### Validating webhook (`/`)

Rejects `Pod` creates/updates where any container (or init container) is
missing CPU/memory resource `requests` or `limits`. Also blocks use of the
`CriticalAddonsOnly` toleration.

## Project layout

```text
webhook/
  main.go           # HTTPS server entrypoint, wires up handlers
  certreloader.go   # Watches /certs/server.crt and /certs/server.key for changes
  mutate/
    handler.go      # AdmissionReview decode/encode for the mutating webhook
    label.go        # environment=default label patch logic
  validate/
    handler.go      # AdmissionReview decode/encode for the validating webhook
    pod.go          # Pod resource request/limit and toleration rules
deployment/
  00-certificates.yaml              # cert-manager Issuer/Certificate resources
  01-service.yaml                   # ClusterIP service fronting the webhook
  02-deployment.yaml                # Webhook Deployment
  03-validatingwebhookconfiguration.yaml
  04-mutatingwebhookconfiguration.yaml
```

## Requirements

- Go 1.27+
- [cert-manager](https://cert-manager.io) installed in the cluster (used to
  issue and inject the webhook's serving certificate)

## Building

```bash
GOFLAGS=-mod=vendor go build ./...
```

## Container image

```bash
docker build -t msp-webhook:1.0 -f dockerfile .
```

The webhook listens on port `9443` and expects its TLS certificate and key to
be mounted at `/certs/server.crt` and `/certs/server.key`.

## Deploying

Apply the manifests in order into the `webhook-system` namespace:

```bash
kubectl apply -f deployment/00-certificates.yaml
kubectl apply -f deployment/01-service.yaml
kubectl apply -f deployment/02-deployment.yaml
kubectl apply -f deployment/03-validatingwebhookconfiguration.yaml
kubectl apply -f deployment/04-mutatingwebhookconfiguration.yaml
```

The `cert-manager.io/inject-ca-from` annotation on the webhook configurations
lets cert-manager's CA injector populate `caBundle` automatically.
