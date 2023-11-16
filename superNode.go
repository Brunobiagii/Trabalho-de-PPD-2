package main

import (
	"context"
	"fmt"
	"flag"
	"bufio"
	"log"
	"strconv"
	"strings"
	"crypto/sha1"
	//Libp2p
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/multiformats/go-multiaddr"
)
//Guardar informações sobre outros nós
type serverNode struct{
	addr   string
	stream network.Stream
	rw     *bufio.ReadWriter
	conectado bool
}


type superNode struct{
	addr   string
	stream network.Stream
	rw     *bufio.ReadWriter
	serverNodes map[int]serverNodes
	conectado bool
}


var ID                int
var superNodes        map[int]superNode
var servidorTerminado bool
var isPrinc           bool
var servRec           int
var servID            int
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

	//Caso não aja destinatário retorna
	if *dest == "" {
		fmt.Println("Informe o destinatário")
		
	} else{  //Caso aja destinatário se conecta à ele

		host := makeHost(*port + 1) //Cria o host
		fmt.Printf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", *port, host.ID())

		rw, err := startAndConnect(ctx, host, *dest, *port)  //Se conecta ao outro host
		if err != nil {
			log.Println(err)
			return
		}
		servidorTerminado = false
		isPrinc = true
		servRec = 0
		startHost(ctx, host, streamHandler)
		select {} //Loop infinito
	}
	
}

func startHost(ctx context.Context, host host.Host, streamHandler network.StreamHandler){
	host.SetStreamHandler("/ola/1.0.0", streamHandler)
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

	rw.WriteString(fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", port, host.ID())) //Manda uma mensagem pela stream para o outro host
	rw.Flush()
	//Lê a resposta do mestre
	str, _ := rw.ReadString('\n') 
	aux := strings.Split(str, ":")
	ID, err = strconv.Atoi(aux[1][:len(aux[1])-1])
	superNodes[ID].addr = fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s\n", port, host.ID())
	superNodes[ID].conectado = true
	servID = ID * 2
	fmt.Println(aux[1], ID)
	//Devolve o ACK
	rw.WriteString(fmt.Sprintf("ACK\n"))
	rw.Flush()
	//Lê quando estiver finalizado
	str, _ = rw.ReadString('\n')
	if str != "\n" {
		fmt.Println(str)
		//Pede informação sobre outros nós
		rw.WriteString(fmt.Sprintf("%v:Roteamento\n", ID))
		rw.Flush()
		str1, _ := rw.ReadString('\n')
		aux = strings.Split(str1, "|")
		fmt.Println(aux)
		for _, it := range aux {
			straux = strings.Split(it, ":")
			idaux, err := strconv.Atoi(straux[0])
			superNodes[idaux].addr = append(superNodesAddr, it)
			superNodes[idaux].conectado = false
		}
		fmt.Println(superNodesAddr)
	}

	return rw, nil
}

//Função que irá ser chamada quando o host for conectado a uma stream
func streamHandler(stream network.Stream) {
	// Cria uma buffered stream para que ler e escrever sejam não bloqueantes
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	str, _ := rw.ReadString("\n")
	splits := strings.Split(str, ":")
	switch splits[0] {
	case "server":
		//hash := sha1.New()
		//SHA-1 Para o nó servidor
		
		//Mensagem para servidor
		msg := fmt.Sprintf("%v:sha1\n", servID)
		superNodes[ID].serverNodes[servID].addr := fmt.Sprintf("%s", str)
		rw.WriteString(msg)
		msg := fmt.Sprintf("%v:NewServer:%s\n", servID, str)
		servID += 1
		str, _ := rw.ReadString("\n")
		if str == "ACK\n" {	
			//Broadcast informações
			
			broadCast(stm.Sprintf(msg))
		}
	case "super":
		idaux, _ := strconv.Atoi(splits[1][:len(splits[1])-1])
		superNodes[idaux].stream = stream
		superNodes[idaux].rw = rw
		superNodes[idaux].conectado = true
	}


	go readStream(rw)
}

func readStream(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString("\n")
		if str == "Terminado\n" {
			servidorTerminado = true
			return
		}
		splits := strings.Split(str, ":")
		id, _ := strconv.Atoi(splits[0])

		switch splits[1] {
		case "NewServer":
			serverNode[id] := splits[2]
			if id < ID {
				isPrinc = false
			}
			servRec += 1
			if servRec >= 5 && isPrinc {
				broadCast("Terminado\n")
			}
		
		case "DelSever":
			//deletar
		}
	}
}

func broadCast(msg string) {
	for i := 0; i < len(superNodes); i++ {
		superNodes[i].rw.WriteString(msg)
		superNodes[i].rw.Flush()
	}
}