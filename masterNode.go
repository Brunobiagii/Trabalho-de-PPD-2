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
	"github.com/libp2p/go-libp2p/core/network"
	//"github.com/libp2p/go-libp2p/core/peer"
	//"github.com/libp2p/go-libp2p/core/peerstore"

	//"github.com/multiformats/go-multiaddr"
)
type superNode struct{
	addr   string
	stream network.Stream
	rw     *bufio.ReadWriter
	registrado bool
}

var superNodes []superNode
var superNodeID int

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
	superNodeID = 0
	host := makeHost(*port) //Cria o host
	log.Printf("use '-d /ip4/127.0.0.1/tcp/%v/p2p/%s' para se conectar a esse host", *port, host.ID())
	startHost(ctx, host, streamHandler)  //Deixa o host esperando conexão
	supCadastrados := false
	for {
		//Verifica se todos os nós forão cadastrados
		if len(superNodes) == 2 && !supCadastrados {
			k := true
			for i := 0; i < len(superNodes); i++ {
				if !superNodes[i].registrado {
					k = false
				}
				
			}
			fmt.Println("Olhado\n")
			//se foram faz um broadcast
			if k {
				fmt.Println("Entrado\n")
				for i := 0; i < len(superNodes); i++ {
					superNodes[i].rw.WriteString(fmt.Sprintf("Terminado\n"))
					superNodes[i].rw.Flush()
				}
				supCadastrados = true
			}
		}
	}
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
	//Cadastra um novo super nó
	if str != "\n"{
		supNodes.addr = str
		supNodes.stream = stream
		supNodes.rw = rw
		supNodes.registrado = false
		superNodes = append(superNodes, supNodes)
		//fmt.Println(str)
	}
	//Retorna o id do super nó
	rw.WriteString(fmt.Sprintf("ID:%v\n", superNodeID))
	rw.Flush()
	id := superNodeID
	superNodeID += 1
	str, _ = rw.ReadString('\n') //Recebe uma mensagem da stream
	fmt.Println(str)
	//Recebe o ack
	if str == "ACK\n" {
		superNodes[id].registrado = true
		fmt.Println("ACK recebido\n")
	}
	//TODO: adicionar o readStream
}

//Receberá as mensagens da stream
func readStream(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			ret := ""
			//Separa a entrada por ":"
			//0 = id, 1 = protocolo, 2 = informação adicional
			splits := strings.Split(str, ":")
			id, _ := strconv.Atoi(splits[0])
			switch splits[1] {
				//Retorna informação de roteamento dos outros super nós
			case "Roteamento":
				for i := 0; i < len(superNodes); i++ {
					if i != id {
						ret = ret + fmt.Sprintf("%v:%s\n", i, superNodes[id].addr)
					}
				}
				superNodes[id].rw.WriteString(ret)
			}
		}
	}
}

//func writeStream(rw *bufio.ReadWriter) { }