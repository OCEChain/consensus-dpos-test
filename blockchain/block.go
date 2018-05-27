package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	)

// Block represents each 'item' in the blockchain
type Block struct {
	Height    	int64
	Timestamp 	int64
	Amount    	int64
	Hash      	string
	PrevHash  	string
	GeneratedBy string
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}