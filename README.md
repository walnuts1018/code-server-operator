# code-server-operator

// TODO(user): Add simple overview of use/purpose

## Description

// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

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
