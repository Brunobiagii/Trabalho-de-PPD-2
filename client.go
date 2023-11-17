package main

import (
	"context"
	"fmt"
	"flag"
	"bufio"
	"log"
	"strconv"
	"strings"
	//Libp2p
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	//"github.com/libp2p/go-libp2p/core/network"
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
	//Flags da linha de comando
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	ctx := context.Background()  //Contexto
	port := flag.Int("p", 8080, "porta")  //Porta do host
	dest := flag.String("d", "", "destinatário")  //Destinatário da conexão
	flag.Parse()

	//Caso não aja destinatário retorna
	if *dest == "" {
		fmt.Println("Informe o destinatário")
		
	} else{  //Caso aja destinatário se conecta à ele

		host := makeHost(*port + 1) //Cria o host
		fmt.Printf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", *port, host.ID())

		_, err := startAndConnect(ctx, host, *dest, *port)  //Se conecta ao outro host
		if err != nil {
			log.Println(err)
			return
		}
		select {} //Loop infinito
	}
}

//Cria uma stream entre os dois hosts
func startAndConnect(ctx context.Context, host host.Host, destination string, port int) (*bufio.ReadWriter, error) {
	//Gera o multi address do destinatário
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Pega as informações do endereço do destinatário
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//Adiciona o destinatário a lista de peers do host
	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	//Cria a stream entre ambos os hosts
	stream, err := host.NewStream(context.Background(), info.ID, "/ola/1.0.0")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Established connection to destination")

	// Cria uma buffered stream para que ler e escrever sejam não bloqueantes
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	rw.WriteString(fmt.Sprintf("client:/ip4/127.0.0.1/tcp/%v/p2p/%s\n", port, host.ID())) //Manda uma mensagem pela stream para o outro host
	rw.Flush()
	//Lê a resposta do mestre
	str, _ := rw.ReadString('\n') 
	aux := strings.Split(str, ":")
	ID, err := strconv.Atoi(aux[1][:len(aux[1])-1])
	fmt.Println(aux[1], ID)
	//Devolve o ACK
	rw.WriteString(fmt.Sprintf("ACK\n"))
	rw.Flush()

	return rw, nil
}