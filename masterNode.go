package main

import (
	//"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
)

func main() {
	master_node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/8080"))
	if err != nil {
		panic(err)
	}

	fmt.Println("master ID: ", master_node.ID())
	for _, addr := range master_node.Addrs() {
		fmt.Printf("%s\n", addr.String())
	}
}
