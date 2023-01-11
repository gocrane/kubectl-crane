# kubectl-crane

[![Go Report Card](https://goreportcard.com/badge/github.com/gocrane/kubectl-crane)](https://goreportcard.com/report/github.com/gocrane/kubectl-crane)

Kubectl plugin for crane, including recommendation and cost estimate.

## Installation 

Downloaded a tar file from [released packages](https://github.com/gocrane/kubectl-crane/releases) and extract `kubectl-crane` from it, then put the binary under your path.

### For Linux 

```bash
export release=v0.2.0
export arch=x86_64
curl -L -o kubectl-crane.tar.gz https://github.com/gocrane/kubectl-crane/releases/download/${release}/kubectl-crane_${release}_Linux_${arch}.tar.gz
tar -xvf kubectl-crane.tar.gz 
cp kubectl-crane_${release}_Linux_${arch}/kubectl-crane /usr/local/bin/
```

### For Mac

```bash
export release=v0.2.0
export arch=arm64
curl -L -o kubectl-crane.tar.gz https://github.com/gocrane/kubectl-crane/releases/download/${release}/kubectl-crane_${release}_Darwin_${arch}.tar.gz
tar -xvf kubectl-crane.tar.gz 
cp kubectl-crane_${release}_Darwin_${arch}/kubectl-crane /usr/local/bin/
```


