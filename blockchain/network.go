package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"io"
	"io/ioutil"
	"time"
	"fmt"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var knownNodes = []string{"localhost:3000"}
var nodeAddress string

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type delegateSt struct {
	AddrFrom string
	Delegates []byte
}

type getblocks struct {
	AddrFrom string
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func sendDelegates(bc *Blockchain, numberDelegate int, delegate *Delegates) {
	//lastBlock := bc.GetLastBlock()
	listDelegates := GetDelegates(bc)
	//tmpDelegate := &Delegates{nodeVersion, lastBlock.Height, nodeAddress, len(listDelegates)}
	for _, tmpDelegate := range listDelegates {
		log.Println(tmpDelegate.Address, nodeAddress, numberDelegate, delegate.NumPeer)
		//send data to all delegate available
		if tmpDelegate.Address != nodeAddress && numberDelegate > delegate.NumPeer{
			data := delegateSt{nodeAddress, delegate.SerializeDelegate()}
			payload := gobEncode(data)
			request := append(commandToBytes("delegates"), payload...)
			sendData(tmpDelegate.Address, request)
		}

		if tmpDelegate.Address != nodeAddress {
			data := delegateSt{tmpDelegate.Address, tmpDelegate.SerializeDelegate()}
			payload := gobEncode(data)
			request := append(commandToBytes("delegates"), payload...)
			sendData(delegate.Address, request)
		}
	}
}

func SendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	log.Println("Recevied a new block!")
	bc.AddBlock(block)
	log.Println("Added block", block.Hash)

	listPeer := GetDelegates(bc)
	for _, peer := range listPeer {
		UpdateDelegate(bc, peer.Address, block.Height)
	}
}

func handleDeletes(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload delegateSt

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	delegateData := payload.Delegates
	delegate := DeserializePeer(delegateData)

	myBestHeight := bc.GetBestHeight()
	// insert delegates
	isInsert := InsertDelegates(bc, delegate, myBestHeight)
	log.Println("handleDelegate", payload.AddrFrom, delegate, myBestHeight, isInsert)

	numberDelegates := GetNumberDelegates(bc)


	//if myBestHeight > delegate.LastHeight {
	//	sendDelegates(bc)
	//}

	log.Println(numberDelegates, delegate.NumPeer)
	//if numberDelegates > delegate.NumPeer {
	if isInsert {
		sendDelegates(bc, numberDelegates, delegate)
	}
	//}


}

func handleConnection(conn net.Conn, bc *Blockchain) {
	time.Sleep(time.Second)
	request, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	log.Printf("Received %s command\n", command)

	switch command {
	case "block":
		handleBlock(request, bc)
	case "delegates":
		handleDeletes(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

// StartServer starts a node
func StartServer(nodeID string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	log.Println("Peer start address: ", nodeAddress)
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockchain(nodeID)

	lastHeight := bc.GetBestHeight()
	numberDelegate := GetNumberDelegates(bc)
	delegate := &Delegates{nodeVersion, lastHeight, nodeAddress, numberDelegate}
	InsertDelegates(bc, delegate, lastHeight)
	if nodeAddress != knownNodes[0] {
		sendDelegates(bc, numberDelegate + 1, delegate)
	}

	go Forks(bc)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}