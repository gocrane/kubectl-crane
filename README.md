# kubectl-crane

[![Go Report Card](https://goreportcard.com/badge/github.com/gocrane/kubectl-crane)](https://goreportcard.com/report/github.com/gocrane/kubectl-crane)

Kubectl plugin for crane, including recommendation and cost estimate.

## Installation

You can install `kubectl-crane` plugin in any of the following ways:

- One-click installation.
- Install using Krew.
- Build from source code.

### One-click installation

Downloaded a tar file from [released packages](https://github.com/gocrane/kubectl-crane/releases) and extract `kubectl-crane` from it, then put the binary under your path.

#### For Linux

```shell
export release=v0.2.0
export arch=x86_64
curl -L -o kubectl-crane.tar.gz https://github.com/gocrane/kubectl-crane/releases/download/${release}/kubectl-crane_${release}_Linux_${arch}.tar.gz
tar -xvf kubectl-crane.tar.gz 
cp kubectl-crane_${release}_Linux_${arch}/kubectl-crane /usr/local/bin/
```

#### For Mac

```shell
export release=v0.2.0
export arch=arm64
curl -L -o kubectl-crane.tar.gz https://github.com/gocrane/kubectl-crane/releases/download/${release}/kubectl-crane_${release}_Darwin_${arch}.tar.gz
tar -xvf kubectl-crane.tar.gz 
cp kubectl-crane_${release}_Darwin_${arch}/kubectl-crane /usr/local/bin/
```

### Install using Krew

`Krew` is the plugin manager for `kubectl` command-line tool.

[Install and setup](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) Krew on your machine.

Then install `kubectl-crane` plug-in:

```shell
kubectl krew install crane
```

### Build from source code

```shell
git clone https://github.com/gocrane/kubectl-crane.git
cd kubectl-crane
export CGO_ENABLED=0
go mod vendor
go build -o kubectl-crane ./cmd/
```

Next, move the `kubectl-crane` executable file in the project root directory to the `PATH` path.

## Usage

```text
$ kubectl crane -h
Kubectl plugin for crane, including recommendation and cost estimate.

Usage:
  kubectl crane [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command
  recommend          view or adopt recommend result
  recommendationrule manage recommendation rules
  version            Print kubectl-crane version
  view-recommend     View a source which recommends related.
```
