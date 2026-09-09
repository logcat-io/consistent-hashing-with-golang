package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

func hash64(s string) uint64 {
	sum := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint64(sum[:8])
}

func owner(key string, nodes []string) string {
	return nodes[hash64(key)%uint64(len(nodes))]
}

func main() {
	before := []string{"a", "b", "c"}
	after := []string{"a", "b", "c", "d"}
	moved := 0
	const keys = 100_000
	for i := 0; i < keys; i++ {
		k := fmt.Sprintf("%d", i)
		if owner(k, before) != owner(k, after) {
			moved++
		}
	}
	fmt.Printf("Moved: %d/%d = %.2f%%\n", moved, keys, float64(moved)*100.0/keys)
}
