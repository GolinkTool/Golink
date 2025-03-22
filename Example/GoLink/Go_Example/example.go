package Go_Example

import (
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

func EthereumSum256(data []byte) (digest [32]byte) {
	h := sha3.NewKeccak256()
	h.Write(data)
	h.Sum(digest[:0])
	return
}
