# `kubectl traceroute`

A `kubectl` plugin to diagnose a service that isn't responding or behaving properly. This plugin attempts to automate to recommendations at https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/.

## Quick Start

```
kubectl traceroute <serviceName><:port>
```

## Example

```
kubectl traceroute -n sentry-enterprise sentry:9000

Tracing route to sentry.sentry-enterprise.svc.cluster.local
  ✓    ✓    ✓    service named sentry found in sentry-enterprise namespace
  ✓    ✓    ✓    port 9000 found on service sentry
  ✓    ✓    ✓    Deployment/sentry-worker
  ✓    ✓    ✓    2 replicas of deployment should be present
 2/2  2/2  2/2   ready replicas of deployment found

```



