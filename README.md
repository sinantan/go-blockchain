# Go Blockchain

A minimal blockchain implementation in Go for learning blockchain basics.

## What is this?

This project shows how blockchain technology works using simple Go code. It's for learning, not for real use.

## Features

- Basic block structure with hash, timestamp, and data
- SHA-256 hash calculation
- Proof-of-Work mining with adjustable difficulty
- Chain validation
- REST API endpoints
- Concurrent mining using goroutines
- No external dependencies (standard library only)

## How to run

You need Go 1.21 or newer.

```bash
git clone https://github.com/sinantan/go-blockchain
cd go-blockchain
go run main.go
```

The server starts on port 8080 with some demo blocks.

## API Endpoints

### View the blockchain
```bash
curl http://localhost:8080/chain
```

### Mine a new block
```bash
curl -X POST http://localhost:8080/mine \
  -H "Content-Type: application/json" \
  -d '{"data":"Alice sends 5 coins to Bob"}'
```

### Check sync status
```bash
curl http://localhost:8080/sync
```

### Change mining difficulty
```bash
curl -X POST http://localhost:8080/difficulty \
  -H "Content-Type: application/json" \
  -d '{"difficulty": 3}'
```

## How it works

1. **Blocks**: Each block has an index, timestamp, data, previous hash, and its own hash
2. **Hashing**: Uses SHA-256 to create unique fingerprints for each block
3. **Mining**: Finds a hash that starts with zeros (proof-of-work)
4. **Chain**: Blocks link together using hash references
5. **Validation**: Checks if all blocks are properly connected and valid



## Testing

Run the program and try:
```bash
curl http://localhost:8080/chain
curl -X POST http://localhost:8080/mine -H "Content-Type: application/json" -d '{"data":"test"}'
```

Watch the mining process in the console output.


## License

MIT License - feel free to use this code for learning.
