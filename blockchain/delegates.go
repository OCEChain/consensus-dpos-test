package blockchain

import (
	"github.com/boltdb/bolt"
	"bytes"
	"encoding/gob"
	"log"
	"errors"
	)

type Delegates struct {
	Version    int64
	LastHeight int64
	Address    string
	NumPeer	   int
}

const peerBucket = "peerBucket"

func GetDelegates(bc *Blockchain) []*Delegates {
	var listDelegate []*Delegates
	bc.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(peerBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			delegate := DeserializePeer(v)
			listDelegate = append(listDelegate, delegate)
		}

		return nil
	})
	return listDelegate
}

// get number delegates
func GetNumberDelegates(bc *Blockchain) int {
	numberDelegate := 0
	bc.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(peerBucket))
		c := b.Cursor()


		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			numberDelegate += 1
		}

		return nil
	})
	return numberDelegate
}


func UpdateDelegate(bc *Blockchain, address string, lastHeight int64)  {
	var delegate Delegates
	bc.db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(peerBucket))
		delegateData := b.Get([]byte(address))
		if delegateData == nil {
			return errors.New("Delegates is not found.")
		}
		delegate = *DeserializePeer(delegateData)
		if delegate.LastHeight < lastHeight {
			delegate.LastHeight = lastHeight
			b.Put([]byte(address), delegate.SerializeDelegate())
			log.Println("updated", address, lastHeight, delegate)
		}
		return nil
	})
}

func InsertDelegates(bc *Blockchain, delegate *Delegates, lastHeight int64) bool{
	isInsert := false
	err := bc.db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(peerBucket))

		delegateData := b.Get([]byte(delegate.Address))
		if delegateData == nil {
			if delegate.LastHeight < lastHeight {
				delegate.LastHeight = lastHeight
			}
			err := b.Put([]byte(delegate.Address), delegate.SerializeDelegate())
			if err != nil {
				log.Panic(err)
			}
			isInsert = true
			return err
		} else {
			tmpDelegate := *DeserializePeer(delegateData)
			if tmpDelegate.LastHeight < lastHeight {
				delegate.LastHeight = lastHeight
				err := b.Put([]byte(delegate.Address), delegate.SerializeDelegate())
				if err != nil {
					log.Panic(err)
				}
				isInsert = true
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return isInsert
}

// Serialize serializes the block
func (b *Delegates) SerializeDelegate() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializePeer(d []byte) *Delegates {
	var delegate Delegates

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&delegate)
	if err != nil {
		log.Panic(err)
	}

	return &delegate
}
