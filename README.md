# os-autoscaler-nodecleaner

Deletes kubernetes nodes in NotReady state by comparing to
the magnum cluster heat stack's list of removed resources.

## Deployment

Uses the same `os-autoscaler-cloud-config` secret from the autoscaler.

Example deployment yaml needs the cluster name to be changed to match
the cluster that is being deployed into.

```
$ kubectl create -f svcaccount.yaml -f nodecleaner.yaml
```
