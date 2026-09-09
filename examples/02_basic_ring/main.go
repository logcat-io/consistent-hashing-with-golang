package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
)

type Node struct {
	Hash uint64
	Name string
}

func h(s string) uint64 {
	sum := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint64(sum[:8])
}

func main() {
	ring := []Node{
		{h("a"), "a"},
		{h("b"), "b"},
		{h("c"), "c"},
	}
	sort.Slice(ring, func(i, j int) bool {
		return ring[i].Hash < ring[j].Hash
	})
	for _, k := range []string{"user:1001", "user:1002", "user:1003"} {
		kh := h(k)
		i := sort.Search(len(ring), func(i int) bool {
			return ring[i].Hash >= kh
		})
		if i == len(ring) {
			i = 0
		}
		fmt.Println(k, "->", ring[i].Name)
	}
}
