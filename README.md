# code-server-operator

code-server-operator は、[coder/code-server](https://github.com/coder/code-server) を Kubernetes 上で管理するための Operator です。
同じ環境が複数台必要なとき、この Operator を使うことで簡単にデプロイすることができます。

## Features

- `CodeServer`リソースを作成することで、Deployment、Service、Ingress、Secret、PVC が作成され、code-server がデプロイされます。
- `CodeServerDeployment`リソースを作成することで、`spec.replicas`に設定した数だけ `CodeServer`リソースが作成され、複数の code-server をデプロイすることができます。

## Install

### Cert-Manager

```shell
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true
```

### Code-Server-Operator

```shell
helm repo add code-server-operator https://walnuts1018.github.io/code-server-operator
helm repo update
helm install code-server-operator code-server-operator/code-server-operator --set fullnameOverride=code-server-operator
```

## Sample

```yaml
apiVersion: cs.walnuts.dev/v1alpha2
kind: CodeServerDeployment
metadata:
  labels:
    app.kubernetes.io/name: codebox
  name: test
spec:
  replicas: 3
  template:
    spec:
      storageSize: 3Gi
      storageClassName: local-path
      initPlugins:
        git:
          repourl: "github.com/walnuts1018/http-dump"
          branch: "master"
        copyDefaultConfig: {}
        copyHome: {}
      envs:
        - name: LANGUAGE_DEFAULT
          value: "ja"
      image: "ghcr.io/coder/code-server:4.89.1"
      domain: "walnuts.dev"
      ingressClassName: "nginx"
```

## Development

### Prerequisites

- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.
- aqua version 2.25.1+

### Install Dependencies

```shell
aqua i
```

### Start Cluster

```shell
make start
tilt up --host 0.0.0.0
```

### Stop Cluster

```shell
make stop
```
