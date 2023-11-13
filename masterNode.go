package main

import (
	"context"
	"fmt"
	"flag"
	"bufio"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)

func makeHost(port int) host.Host {
	host, err := libp2p.New(libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)))
	if err != nil {
		panic(err)
	}
	return host
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	ctx := context.Background()
	port := flag.Int("p", 8080, "porta")
	dest := flag.String("d", "", "destinatário")

	if *dest == "" {
		host := makeHost(*port)
		log.Printf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", *port, host.ID())
		startHost(ctx, host, streamHandler)
		
	} else{

		host := makeHost(*port + 1)

		fmt.Println("ID: ", host.ID())
		for _, addr := range host.Addrs() {
			fmt.Printf("%s\n", addr.String())
		}

		rw, err := startAndConnect(ctx, host, *dest)
		if err != nil {
			log.Println(err)
			return
		}
		rw.WriteString(fmt.Sprintf("Olhhaaa\n"))
		rw.Flush()
	}
	select {}
}

func startHost(ctx context.Context, host host.Host, streamHandler network.StreamHandler){
	host.SetStreamHandler("/ola/1.0.0", streamHandler)
}

func startAndConnect(ctx context.Context, host host.Host, destination string) (*bufio.ReadWriter, error) {
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	s, err := host.NewStream(context.Background(), info.ID, "/ola/1.0.0")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Established connection to destination")

	// Create a buffered stream so that read and writes are non-blocking.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	return rw, nil
}

func streamHandler(stream network.Stream) {
	defer stream.Close()

	buf := bufio.NewReader(stream)
	fmt.Println(buf)
	
}