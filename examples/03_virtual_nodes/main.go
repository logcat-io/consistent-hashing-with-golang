package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
)

type V struct {
	Hash uint64
	Node string
}

func h(s string) uint64 {
	sum := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint64(sum[:8])
}

func main() {
	var ring []V
	for _, n := range []string{"a", "b", "c"} {
		for i := 0; i < 100; i++ {
			ring = append(ring, V{h(fmt.Sprintf("%s#%d", n, i)), n})
		}
	}
	sort.Slice(ring, func(i, j int) bool {
		return ring[i].Hash < ring[j].Hash
	})
	counts := map[string]int{}
	for i := 0; i < 100; i++ {
		kh := h(fmt.Sprintf("user:%d", i))
		x := sort.Search(len(ring), func(j int) bool {
			return ring[j].Hash >= kh
		})
		if x == len(ring) {
			x = 0
		}
		counts[ring[x].Node]++
	}
	fmt.Println("counts:", counts)
}
