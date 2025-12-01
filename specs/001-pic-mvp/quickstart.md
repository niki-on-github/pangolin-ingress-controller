# Quickstart: Pangolin Ingress Controller MVP

## Prerequisites

1. **Kubernetes cluster** (1.28+)
2. **pangolin-operator** installed with CRDs
3. **At least one PangolinTunnel** configured and ready
4. **kubectl** configured

## Installation

### 1. Deploy PIC

```bash
# Apply RBAC and Deployment
kubectl apply -f https://raw.githubusercontent.com/<repo>/main/deploy/install.yaml

# Or using kustomize
kubectl apply -k config/default
```

### 2. Verify Installation

```bash
# Check controller is running
kubectl get pods -n pangolin-system -l app=pangolin-ingress-controller

# Check logs
kubectl logs -n pangolin-system -l app=pangolin-ingress-controller
```

## Usage

### Expose a Service

```yaml
# Create an Ingress with ingressClassName: pangolin
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  namespace: default
spec:
  ingressClassName: pangolin
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app-service
                port:
                  number: 8080
```

```bash
kubectl apply -f ingress.yaml
```

### Verify Exposure

```bash
# Check PangolinResource was created
kubectl get pangolinresources -n default

# Check events on Ingress
kubectl describe ingress my-app

# Expected event:
# Normal  Created  <time>  pangolin-ingress-controller  Created PangolinResource pic-default-my-app-<hash>
```

### Update Configuration

```yaml
# Update host
kubectl patch ingress my-app --type=merge -p '{"spec":{"rules":[{"host":"newapp.example.com","http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"my-app-service","port":{"number":8080}}}}]}}]}}'
```

### Disable Temporarily

```bash
# Add annotation to disable
kubectl annotate ingress my-app pangolin.ingress.k8s.io/enabled=false

# Re-enable
kubectl annotate ingress my-app pangolin.ingress.k8s.io/enabled-
```

### Remove Exposure

```bash
# Delete Ingress (PangolinResource deleted automatically)
kubectl delete ingress my-app
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PIC_DEFAULT_TUNNEL_NAME` | `default` | Default tunnel for `ingressClassName: pangolin` |
| `PIC_TUNNEL_CLASS_MAPPING` | `` | Map of class suffix to tunnel name |
| `PIC_BACKEND_SCHEME` | `http` | Backend protocol |
| `PIC_RESYNC_PERIOD` | `5m` | Controller resync interval |
| `PIC_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `PIC_WATCH_NAMESPACES` | `` | Comma-separated namespaces (empty=all) |

### Multi-Tunnel Setup

```yaml
# Set tunnel mapping in Deployment
env:
  - name: PIC_TUNNEL_CLASS_MAPPING
    value: |
      eu=tunnel-eu
      us=tunnel-us
```

```yaml
# Use specific tunnel via ingressClassName
spec:
  ingressClassName: pangolin-eu  # Uses tunnel-eu
```

## Troubleshooting

### PangolinResource Not Created

```bash
# Check controller logs
kubectl logs -n pangolin-system -l app=pangolin-ingress-controller

# Check Ingress events
kubectl describe ingress <name>

# Common causes:
# - ingressClassName not "pangolin" or "pangolin-*"
# - Tunnel not found
# - Invalid host format
```

### Resource Not Syncing

```bash
# Check PangolinResource status
kubectl get pangolinresource <name> -o yaml

# Check pangolin-operator logs
kubectl logs -n pangolin-system -l app=pangolin-operator
```

## Validation Checklist

- [ ] Controller pod is Running
- [ ] PangolinTunnel exists and is Ready
- [ ] Ingress has correct `ingressClassName`
- [ ] Host is valid (no wildcards, no IPs)
- [ ] Path is `/` (MVP limitation)
- [ ] Backend service exists
