package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultDifficulty = 2
	MiningPort        = ":8080"
)

type Block struct {
	Index        int    `json:"index"`
	Timestamp    int64  `json:"timestamp"`
	Data         string `json:"data"`
	PreviousHash string `json:"previous_hash"`
	Hash         string `json:"hash"`
	Nonce        int    `json:"nonce"`
	Difficulty   int    `json:"difficulty"`
}

type Blockchain struct {
	Chain      []Block `json:"chain"`
	Difficulty int     `json:"difficulty"`
	mutex      sync.RWMutex
}

var blockchain *Blockchain

func NewBlock(index int, data string, previousHash string, difficulty int) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: previousHash,
		Difficulty:   difficulty,
		Nonce:        0,
	}

	block.Hash = block.mineBlock()
	return block
}

func (b *Block) calculateHash() string {
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) +
		b.Data + b.PreviousHash + strconv.Itoa(b.Nonce)

	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func (b *Block) mineBlock() string {
	target := strings.Repeat("0", b.Difficulty)

	fmt.Printf("Mining block with difficulty %d...\n", b.Difficulty)
	start := time.Now()

	for {
		hash := b.calculateHash()

		if strings.HasPrefix(hash, target) {
			fmt.Printf("Block mined! Hash: %s (Time: %v, Nonce: %d)\n",
				hash, time.Since(start), b.Nonce)
			return hash
		}

		b.Nonce++
	}
}

func NewBlockchain() *Blockchain {
	genesisBlock := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Data:         "Genesis Block",
		PreviousHash: "",
		Hash:         "",
		Nonce:        0,
		Difficulty:   DefaultDifficulty,
	}

	genesisBlock.Hash = genesisBlock.calculateHash()

	return &Blockchain{
		Chain:      []Block{genesisBlock},
		Difficulty: DefaultDifficulty,
	}
}

func (bc *Blockchain) AddBlock(data string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	previousBlock := bc.Chain[len(bc.Chain)-1]
	newIndex := previousBlock.Index + 1

	newBlock := NewBlock(newIndex, data, previousBlock.Hash, bc.Difficulty)
	bc.Chain = append(bc.Chain, *newBlock)
}

func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		if currentBlock.Hash != currentBlock.calculateHash() {
			fmt.Printf("Invalid hash at block %d\n", i)
			return false
		}

		if currentBlock.PreviousHash != previousBlock.Hash {
			fmt.Printf("Invalid previous hash at block %d\n", i)
			return false
		}

		target := strings.Repeat("0", currentBlock.Difficulty)
		if !strings.HasPrefix(currentBlock.Hash, target) {
			fmt.Printf("Invalid proof of work at block %d\n", i)
			return false
		}
	}

	return true
}

func (bc *Blockchain) SetDifficulty(difficulty int) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if difficulty > 0 && difficulty < 10 {
		bc.Difficulty = difficulty
		fmt.Printf("Mining difficulty set to %d\n", difficulty)
	}
}

func getChain(w http.ResponseWriter, r *http.Request) {
	blockchain.mutex.RLock()
	defer blockchain.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"chain":    blockchain.Chain,
		"length":   len(blockchain.Chain),
		"is_valid": blockchain.IsValid(),
	}

	json.NewEncoder(w).Encode(response)
}

func mineBlock(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Data string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if requestData.Data == "" {
		requestData.Data = fmt.Sprintf("Block %d - %s",
			len(blockchain.Chain),
			time.Now().Format("2006-01-02 15:04:05"))
	}

	go func() {
		fmt.Printf("Starting to mine new block: %s\n", requestData.Data)
		blockchain.AddBlock(requestData.Data)
	}()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "Block mining started",
		"data":    requestData.Data,
	}
	json.NewEncoder(w).Encode(response)
}

func syncChain(w http.ResponseWriter, r *http.Request) {
	blockchain.mutex.RLock()
	defer blockchain.mutex.RUnlock()

	peerChains := []map[string]interface{}{
		{
			"peer_id": "node-1",
			"length":  len(blockchain.Chain),
			"status":  "synced",
		},
		{
			"peer_id": "node-2",
			"length":  len(blockchain.Chain) - 1,
			"status":  "syncing",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"current_length": len(blockchain.Chain),
		"peers":          peerChains,
		"sync_status":    "active",
		"is_valid":       blockchain.IsValid(),
	}

	json.NewEncoder(w).Encode(response)
}

func setDifficulty(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Difficulty int `json:"difficulty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	blockchain.SetDifficulty(requestData.Difficulty)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message":    "Difficulty updated",
		"difficulty": blockchain.Difficulty,
	}
	json.NewEncoder(w).Encode(response)
}

func startServer() {
	http.HandleFunc("/chain", getChain)
	http.HandleFunc("/mine", mineBlock)
	http.HandleFunc("/sync", syncChain)
	http.HandleFunc("/difficulty", setDifficulty)

	fmt.Printf("Blockchain API server starting on http://localhost%s\n", MiningPort)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /chain      - View the blockchain")
	fmt.Println("  POST /mine       - Mine a new block")
	fmt.Println("  GET  /sync       - Check peer sync status")
	fmt.Println("  POST /difficulty - Set mining difficulty")

	log.Fatal(http.ListenAndServe(MiningPort, nil))
}

func main() {
	fmt.Println("Starting minimal blockchain...")

	blockchain = NewBlockchain()
	fmt.Printf("Genesis block created: %s\n", blockchain.Chain[0].Hash)

	fmt.Println("Adding demo blocks...")

	var wg sync.WaitGroup

	demoData := []string{
		"Alice sends 10 coins to Bob",
		"Bob sends 5 coins to Charlie",
		"Charlie sends 3 coins to Alice",
	}

	for _, data := range demoData {
		wg.Add(1)
		go func(blockData string) {
			defer wg.Done()
			blockchain.AddBlock(blockData)
		}(data)
	}

	wg.Wait()

	fmt.Printf("Blockchain valid: %v\n", blockchain.IsValid())
	fmt.Printf("Total blocks: %d\n", len(blockchain.Chain))

	startServer()
}
