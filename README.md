# Pangolin Ingress Controller (PIC)

A Kubernetes Ingress Controller that exposes services via [Pangolin](https://github.com/your-org/pangolin) by creating `PangolinResource` CRDs.

## Overview

PIC enables a **Kubernetes-native experience** for exposing services through Pangolin. Instead of manually configuring Pangolin, you simply create a standard Kubernetes `Ingress` resource.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
spec:
  ingressClassName: pangolin
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 8080
```

PIC will automatically create a `PangolinResource` that `pangolin-operator` processes to expose your service.

## Prerequisites

- Kubernetes 1.28+
- `pangolin-operator` installed with CRDs
- At least one `PangolinTunnel` configured

## Installation

### Helm (recommended)

```bash
helm install pic ./charts/pangolin-ingress-controller \
  --namespace pangolin-system \
  --create-namespace
```

### With custom values

```bash
helm install pic ./charts/pangolin-ingress-controller \
  --namespace pangolin-system \
  --create-namespace \
  --set config.defaultTunnelName=my-tunnel \
  --set config.logLevel=debug
```

### Raw manifests

```bash
kubectl apply -f https://raw.githubusercontent.com/wizzz/pangolin-ingress-controller/main/deploy/install.yaml
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PIC_DEFAULT_TUNNEL_NAME` | `default` | Tunnel for `ingressClassName: pangolin` |
| `PIC_TUNNEL_CLASS_MAPPING` | - | Multi-tunnel mapping (see below) |
| `PIC_BACKEND_SCHEME` | `http` | Backend protocol |
| `PIC_RESYNC_PERIOD` | `5m` | Reconciliation interval |
| `PIC_LOG_LEVEL` | `info` | Log level |
| `PIC_WATCH_NAMESPACES` | - | Limit to specific namespaces |

### Multi-Tunnel Setup

```yaml
env:
  - name: PIC_TUNNEL_CLASS_MAPPING
    value: |
      eu=tunnel-eu
      us=tunnel-us
```

Then use `ingressClassName: pangolin-eu` to route through `tunnel-eu`.

### Annotations

| Annotation | Description |
|------------|-------------|
| `pangolin.ingress.k8s.io/enabled` | Enable/disable (`true`/`false`) |
| `pangolin.ingress.k8s.io/tunnel-name` | Override tunnel |
| `pangolin.ingress.k8s.io/domain-name` | Override domain |
| `pangolin.ingress.k8s.io/subdomain` | Override subdomain |

## Development

```bash
# Build
make build

# Test
make test

# Run locally (requires kubeconfig)
make run

# Build Docker image
make docker-build
```

## Architecture

```
Ingress ──▶ PIC ──▶ PangolinResource ──▶ pangolin-operator ──▶ Pangolin API
```

PIC only manages `PangolinResource` objects. All Pangolin API interaction is handled by `pangolin-operator`.

## License

Apache 2.0
