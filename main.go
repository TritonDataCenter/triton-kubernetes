package main

import "github.com/joyent/triton-kubernetes/cmd"

var GitHash string = "unknown commit"

func main() {
        cmd.GitHash = GitHash
	cmd.Execute()
}
