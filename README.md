# About

A demo to show how to develop a golang App regarding the go-clean-arch philosophy, then deploy it with the k8s toolchain.

# Usage

## Setup

- Modify the `hostAliases` section in both k8s-pod.yaml and k8s-metrics.yaml, replace the `ip` field with the actual value on your machine.
- Run Elasticsearch and Kibana (you may need to install them first).
- Start the docker engine.
- Start the minikube.
- Create the secrets needed by your app.
- Choose a dev tool to start.
- Populate the database by using the scripts under `config/migrate`

### Create secrets

```shell
kubectl create secret generic go-boilerplate-secret \
        --from-literal=jwt_secret=your_secret \
        --from-literal=db_password=your_password \
        --from-literal=kv_password=your_password
```

## Tilt

It reads the `Tiltfile` file in the current directory and brings up the services accordingly.

Run `tilt up`

>installation: `curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash`

### Teardown

```shell
$ tilt down
```

The command will trigger the Tilt to delete any k8s resources it allocated before.

## Skaffold

The same with the Tilt but with the remote debugging support.

dev:

```shell
$ skaffold dev --port-forward
```

debug:

```shell
$ skaffold debug --port-forward
```

It reads the service definitions and automatically forwards the ports defined in the `portForward` section. Add your own port-forward rules if necessary.

### Teardown

It automatically cleans the resources upon exit

# Monitor

## Metrics

It's not enabled by default, to enable it, please run:

```shell
$ kubectl apply -f k8s-metrics.yaml
```

Now you can check the metrics data in the Kibana dashboard.