# About

A demo to show how to develope a golang App using go-clean-arch with k8s toolkit.

# Usage

## Setup

- Modify the `hostAliases` section in both k8s-pod.yaml and k8s-metrics.yaml, replace the `ip` field with the actual value on your machine.
- Run Elasticsearch and Kibana (you may need to install them first).
- Start docker engine.
- Start minikube.

## Tilt

It will read the `Tiltfile` file in CWD, and bring up the pods accordingly.

Just run `tilt up`

>You may install it by `curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash`

### Teardown

While the tilt dashboard is still running:

```shell
$ tilt down
```

The command will trigger tilt to delete any k8s resource it created before.

## Skaffold

It got the same functionality as tilt, but with remote debug support.

dev:

```shell
$ skaffold dev --port-forward
```

debug:

```shell
$ skaffold debug --port-forward
```

It will read the k8s service definitions, and automatically forward the port for you. You may specify the custom port-forward rules in `skaffold.yaml`.

### Teardown

It will clean resources upon exit

# Monitor

## Metrics

It's not enabled by default, you may run:

```shell
$ kubectl apply -f k8s-metrics.yaml
```

Now you can view the metrics data in Kibana dashboard