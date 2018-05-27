package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
	"errors"
	"time"
	"encoding/hex"
	"crypto/sha256"
	"fmt"
	"../utils"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"


// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		hash := []byte(block.Hash)
		blockInDb := b.Get(hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("lastHash"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("lastHash"), hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int64 {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("lastHash"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}


// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetLastBlock() Block {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("lastHash"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock
}


// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, []byte(block.Hash))

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}


func setupDB(dbName string) (*bolt.DB, error) {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(peerBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
	return db, nil
}

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block *Block) string {
	record := string(block.Height) + string(block.Timestamp) + string(block.Amount) + block.PrevHash
	return calculateHash(record)
}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, Amount int64, GeneratedBy string) *Block {
	newBlock := &Block{}

	t := time.Now()

	newBlock.Height = oldBlock.Height + 1
	newBlock.Timestamp = t.Unix()
	newBlock.Amount = Amount
	newBlock.GeneratedBy = GeneratedBy
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)

	return newBlock
}

func genBlock(db *bolt.DB) *Block {
	block := &Block{0, time.Now().Unix(), 0, "", "", "root"}
	block.Hash = calculateBlockHash(block)
	return block
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	db, err := setupDB(dbFile)

	var tip []byte
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("lastHash"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	if len(tip) == 0 {
		block := genBlock(db)
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			err = b.Put([]byte("lastHash"), []byte(block.Hash))
			err = b.Put([]byte(block.Hash), block.Serialize())
			return nil
		})
		tip = []byte(block.Hash)
	}
	fmt.Println(tip)
	bc := Blockchain{tip, db}

	return &bc
}

func genBlockPeriod(bc *Blockchain)  {
	time.Sleep(utils.BLOCK_TIME * time.Second)
	listDelegates := GetDelegates(bc)
	if len(listDelegates) > 3 {
		lastBlock := bc.GetLastBlock()
		//timeNow := time.Now().Unix()
		indexDelegate := 0
		delegate := listDelegates[indexDelegate]
		if nodeAddress == knownNodes[0] {
			quorum := 0
			noquorum := 0
			for _, delegatePeer := range listDelegates {
				if delegatePeer.LastHeight == lastBlock.Height {
					quorum += 1
				} else {
					noquorum += 1
				}
			}

			log.Println(quorum, noquorum)

			if (quorum+noquorum) > 0 && float64(float64(quorum)/float64(quorum+noquorum)) > float64(0.66) {
				block := generateBlock(lastBlock, 0, delegate.Address)
				bc.AddBlock(block)

				log.Println(block)
				for _, delegatePeer := range listDelegates {
					if delegatePeer.Address != nodeAddress {
						log.Println("sendBlock", delegatePeer.Address)
						SendBlock(delegatePeer.Address, block)
						UpdateDelegate(bc, delegatePeer.Address, block.Height)
					}
				}
			}
		} else {
			log.Println(delegate.Address)
		}
	} else {
		log.Println("len peer: ", len(listDelegates))
	}
}

// Delegates Proof of State
func Forks(bc *Blockchain)  {
	for {
		genBlockPeriod(bc)
	}
}