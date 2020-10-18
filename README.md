# About

A demo to show how to develope a golang App using go-clean-arch with k8s toolkit.

# Usage

## Setup

- Modify the `hostAliases` section in both k8s-pod.yaml and k8s-metrics.yaml, replace the `ip` field with the actual value on your machine.
- Run Elasticsearch and Kibana (you may need to install them first).
- Start docker engine.
- Start minikube.
- Create secrets.
- Choose a dev tool to start App
- Create database using scripts under config/sql

### Create secrets

```shell
kubectl create secret generic go-boilerplate-secret \
        --from-literal=jwt_secret=your_secret \
        --from-literal=db_password=your_password \
        --from-literal=kv_password=your_password
```

## Tilt

It will read the `Tiltfile` file in current directory and bring up the pods accordingly.

Run `tilt up`

>You may install it by `curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash`

### Teardown

```shell
$ tilt down
```

The command will trigger Tilt to delete any k8s resource it created before.

## Skaffold

The same with tilt but with remote debug support.

dev:

```shell
$ skaffold dev --port-forward
```

debug:

```shell
$ skaffold debug --port-forward
```

It will read the k8s service definitions and automatically forward the port for you, add your own port-forward rules in `skaffold.yaml`.

### Teardown

Automatically clean resources upon exit

# Monitor

## Metrics

It's not enabled by default, to enable it, you may run:

```shell
$ kubectl apply -f k8s-metrics.yaml
```

Now you can view the metrics data in Kibana dashboard