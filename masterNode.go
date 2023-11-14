package main

import (
	"context"
	"fmt"
	"flag"
	"bufio"
	"log"
	//Libp2p
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	//"github.com/libp2p/go-libp2p/core/peer"
	//"github.com/libp2p/go-libp2p/core/peerstore"

	//"github.com/multiformats/go-multiaddr"
)
type superNode struct{
	addr   string
	stream network.Stream
	rw     *bufio.ReadWriter
}

var superNodes []superNode

//Cria um host usando o libp2p com a porta específicada
func makeHost(port int) host.Host {
	host, err := libp2p.New(libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)))
	if err != nil {
		panic(err)
	}
	return host
}

func main() {
	//Flags da linha de comando
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	ctx := context.Background()  //Contexto
	port := flag.Int("p", 8080, "porta")  //Porta do host
	flag.Parse()

	host := makeHost(*port) //Cria o host
	log.Printf("use '-d /ip4/127.0.0.1/tcp/%v/p2p/%s' para se conectar a esse host", *port, host.ID())
	startHost(ctx, host, streamHandler)  //Deixa o host esperando conexão

	select {} //Loop infinito
}

// Inicializa o stream handler do host que irá esperar conexão
func startHost(ctx context.Context, host host.Host, streamHandler network.StreamHandler){
	host.SetStreamHandler("/ola/1.0.0", streamHandler)
}

//Função que irá ser chamada quando o host for conectado a uma stream
func streamHandler(stream network.Stream) {
	var supNodes superNode
	// Cria uma buffered stream para que ler e escrever sejam não bloqueantes
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	str, _ := rw.ReadString('\n') //Recebe uma mensagem da stream
	if str != "\n"{
		supNodes.addr = str
		supNodes.stream = stream
		supNodes.rw = rw
		superNodes = append(superNodes, supNodes)
		//fmt.Println(str)
	}
	for i := 0; i < len(superNodes); i++ {
		fmt.Println(superNodes[i].addr)
	}
}