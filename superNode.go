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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)

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
	dest := flag.String("d", "", "destinatário")  //Destinatário da conexão
	flag.Parse()

	//Caso não aja destinatário o host é criado e espera conexão
	if *dest == "" {
		fmt.Println("Informe o destinatário")
		
	} else{  //Caso aja destinatário se conecta à ele

		host := makeHost(*port + 1) //Cria o host
		fmt.Printf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", *port, host.ID())

		rw, err := startAndConnect(ctx, host, *dest)  //Se conecta ao outro host
		if err != nil {
			log.Println(err)
			return
		}
		rw.WriteString(fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", *port, host.ID())) //Manda uma mensagem pela stream para o outro host
		rw.Flush()
		select {} //Loop infinito
	}
	
}
//Cria uma stream entre os dois hosts
func startAndConnect(ctx context.Context, host host.Host, destination string) (*bufio.ReadWriter, error) {
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

	return rw, nil
}

//Função que irá ser chamada quando o host for conectado a uma stream
func streamHandler(stream network.Stream) {
	defer stream.Close() //fecha a stream ao final da função

	// Cria uma buffered stream para que ler e escrever sejam não bloqueantes
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	str, _ := rw.ReadString('\n') //Recebe uma mensagem da stream
	if str != "\n"{
		fmt.Println(str) //Printa a mensagem
	}
	
	
}