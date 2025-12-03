# Quickstart: SSO & Multi-Path Support

**Feature**: 002-sso-multipath-support  
**Date**: 2025-12-03

## Prerequisites

- Kubernetes cluster with PIC and pangolin-operator deployed
- Pangolin tunnel configured and connected
- Ingress resource with `ingressClassName: pangolin`

## Usage Examples

### 1. Public Service (SSO Disabled - Default)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: public-app
  annotations:
    pangolin.ingress.k8s.io/sso: "false"
    pangolin.ingress.k8s.io/block-access: "false"
spec:
  ingressClassName: pangolin
  rules:
    - host: public.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 8080
```

**Result**: Service accessible without authentication.

### 2. Protected Service (SSO Enabled)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: protected-app
  annotations:
    pangolin.ingress.k8s.io/sso: "true"
    pangolin.ingress.k8s.io/block-access: "true"
spec:
  ingressClassName: pangolin
  rules:
    - host: admin.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: admin-service
                port:
                  number: 8080
```

**Result**: Users must authenticate before accessing the service.

### 3. Multi-Path Routing

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: multi-path-app
  annotations:
    pangolin.ingress.k8s.io/sso: "false"
spec:
  ingressClassName: pangolin
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-service
                port:
                  number: 3000
          - path: /web
            pathType: Prefix
            backend:
              service:
                name: frontend-service
                port:
                  number: 80
          - path: /
            pathType: Prefix
            backend:
              service:
                name: default-service
                port:
                  number: 8080
```

**Result**: 
- `app.example.com/api/*` → api-service:3000
- `app.example.com/web/*` → frontend-service:80
- `app.example.com/*` → default-service:8080

## Annotation Reference

| Annotation | Type | Default | Description |
|------------|------|---------|-------------|
| `pangolin.ingress.k8s.io/sso` | `"true"` / `"false"` | `"false"` | Enable SSO authentication |
| `pangolin.ingress.k8s.io/block-access` | `"true"` / `"false"` | `"false"` | Block unauthenticated access (requires `sso: "true"`) |

## Verification

### Check PangolinResource

```bash
kubectl get pangolinresource -n <namespace> -o yaml
```

Expected fields in `spec.httpConfig`:
```yaml
spec:
  httpConfig:
    domainName: example.com
    subdomain: app
    sso: false
    blockAccess: false
```

Expected fields in `spec.targets` (for multi-path):
```yaml
spec:
  targets:
    - ip: api-service.<namespace>.svc.cluster.local
      port: 3000
      path: /api
      pathMatchType: prefix
      priority: 200
    - ip: frontend-service.<namespace>.svc.cluster.local
      port: 80
      path: /web
      pathMatchType: prefix
      priority: 200
    - ip: default-service.<namespace>.svc.cluster.local
      port: 8080
      path: /
      pathMatchType: prefix
      priority: 100
```

### Check Pangolin Dashboard

1. Navigate to Pangolin dashboard
2. Find the resource under your organization
3. Verify SSO settings in resource configuration
4. Verify targets list shows all paths with correct routing

## Troubleshooting

### SSO Not Applying

1. Check annotation spelling: `pangolin.ingress.k8s.io/sso`
2. Verify annotation value is exactly `"true"` or `"false"` (not `true` without quotes)
3. Check operator logs: `kubectl logs -n <operator-ns> -l app=pangolin-operator`
4. Verify PangolinResource has `sso` field in httpConfig

### Paths Not Routing Correctly

1. Verify all paths are present in PangolinResource `spec.targets`
2. Check priority values (more specific paths should have higher priority)
3. Verify pathMatchType matches Ingress pathType
4. Check target IPs resolve to correct services

### Resource Not Created

1. Check PIC logs: `kubectl logs -n pangolin-system -l app=pangolin-ingress-controller`
2. Verify Ingress has `ingressClassName: pangolin`
3. Check for validation errors in events: `kubectl describe ingress <name>`
