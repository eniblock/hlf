Hyperledger Fabric on Kubernetes

# Helm charts

```
helm install peer1 oci://ghcr.io/eniblock/hlf-peer --version 0.2.0
```

# Development

## Requirements

- [Docker](https://docs.docker.com/engine/install/#server)
- [clk k8s](https://github.com/click-project/clk_recipe_k8s)

Install your local kubernetes cluster with:

```shell script
sudo apt-get install pip
curl -sSL https://clk-project.org/install.sh | env CLK_EXTENSIONS=k8s bash
clk k8s flow
```

## Updates
```shell
clk extension update k8s
clk k8s flow
```

## Start the application

```shell script
tilt up
```

## Stop the application

```shell script
tilt down
```

The option `--no-volumes` can be used to keep the volumes.
