# Building `triton-kubernetes` binary manually

- Install Go
- Set `GOBIN` and `GOPATH`
- Clone this repository into `$GOPATH/src/github.com/joyent/triton-kubernetes`
- Run `go get` and `go install` from that directory

This will build the `triton-kubernetes` binary into `$GOBIN`.



Note: To build the project with an embedded git hash:
```
go build -ldflags "-X main.GitHash=$(git rev-list -1 HEAD)"
```