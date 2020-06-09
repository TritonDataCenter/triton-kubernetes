# Build and install `triton-kubernetes`

## Install Go

- [Download and Install](https://github.com/golang/go#download-and-install)

## Setting `GOPATH`

`GOPATH` can be any directory on your system. In Unix examples, we will set it to `$HOME/go`. Another common setup is to set `GOPATH=$HOME`.

### Bash

Edit your `~/.bash_profile` to add the following line:

```bash
export GOPATH=$HOME/go
```

Save and exit your editor. Then, source your `~/.bash_profile`.

```bash
source ~/.bash_profile
```

Note: Set the GOBIN path to generate a binary file when `go install` is run.

```bash
export GOBIN=$HOME/go/bin
```

### Zsh

Edit your `~/.zshrc` file to add the following line:

```bash
export GOPATH=$HOME/go
```

Save and exit your editor. Then, source your `~/.zshrc`.

```bash
source ~/.zshrc
```

### Build binary via MakeFie

To build binaries for `osx`, `linux`, `linux-rpm` and `linux-debian` under `build` folder, run the following:

```bash
make build
```

### Install and Run

- Clone this repository into `$GOPATH/src/github.com/joyent/triton-kubernetes`
- Run `go get` and `go install` from that directory

This will build the `triton-kubernetes` binary into `$GOBIN`.

You can now run cli in your terminal like below

```bash
triton-kubernetes --help
```

Note: To build the project with an embedded git hash:

```bash
go build -ldflags "-X main.GitHash=$(git rev-list -1 HEAD)"
```
