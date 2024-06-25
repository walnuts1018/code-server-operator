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

## Spec

```go
type CodeServerSpec struct {
    // Specifies the storage size that will be used for code server
    // +kubebuilder:validation:Pattern="^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$"
    // +kubebuilder:default="1Gi"
    StorageSize string `json:"storageSize,omitempty"`

    // Specifies the storage class name for persistent volume claim
    StorageClassName string `json:"storageClassName,omitempty"`

    // Specifies the additional annotations for persistent volume claim
    StorageAnnotations map[string]string `json:"storageAnnotations,omitempty"`

    // VolumeName specifies the volume name for persistent volume claim
    VolumeName string `json:"volumeName,omitempty"`

    // Specifies the resource requirements for code server pod.
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // Specifies the period before controller suspend the resources (delete all resources except data).
    SuspendAfterSeconds *int64 `json:"suspendAfterSeconds,omitempty"`

    // Specifies the domain for code server
    Domain string `json:"domain,omitempty"`

    // Specifies the envs
    Envs []corev1.EnvVar `json:"envs,omitempty"`

    // Specifies the image used to running code server
    // +kubebuilder:default="ghcr.io/coder/code-server:latest"
    Image string `json:"image,omitempty"`

    // Specifies the init plugins that will be running to finish before code server running.
    InitPlugins map[string]map[string]string `json:"initPlugins,omitempty"`

    // Specifies the node selector for scheduling.
    NodeSelector map[string]string `json:"nodeSelector,omitempty"`

    // Specifies the terminal container port for connection, defaults in 19200.
    // +kubebuilder:default=19200
    ContainerPort int32 `json:"containerPort,omitempty"`

    // ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
    ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

    IngressClassName string `json:"ingressClassName,omitempty"`

    // PublicProxyPorts specifies the public proxy ports for code server
    PublicProxyPorts []int32 `json:"publicProxyPorts,omitempty"`

    //InitCommand specifies the init commands that will be running to finish before code server running.
    InitCommand string `json:"initCommand,omitempty"`
}
```

## InitPlugins

```go
type gitPlugin struct {
    Repourl    string `required:"true" json:"repourl"`
    Branch     string `json:"branch"`
}
```

```go
type copyDefaultConfigPlugin struct {
    Image      string `required:"true" json:"image"`
}
```

```go
type copyHomePlugin struct {
    Image      string `required:"true" json:"image"`
}
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
