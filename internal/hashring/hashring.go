package hashring

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type HashFunc func(string) uint64

type Node struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type VirtualNode struct {
	Hash    uint64 `json:"hash"`
	Node    string `json:"node"`
	Replica int    `json:"replica"`
}

type Ring struct {
	mu              sync.RWMutex
	replicasPerUnit int
	hash            HashFunc
	nodes           map[string]Node
	vnodes          []VirtualNode
}

func SHA256Hash64(value string) uint64 {
	sum := sha256.Sum256([]byte(value))
	return binary.BigEndian.Uint64(sum[:8])
}

func New(replicasPrtUnit int, hash HashFunc) *Ring {
	if replicasPrtUnit < 1 {
		replicasPrtUnit = 1
	}
	if hash == nil {
		hash = SHA256Hash64
	}
	return &Ring{
		replicasPerUnit: replicasPrtUnit,
		hash:            hash,
		nodes:           map[string]Node{},
	}
}

func (r *Ring) AddNode(name string, weight int) error {
	if name == "" {
		return errors.New("node name is required")
	}
	if weight < 1 {
		return errors.New("weight must be greater than one")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.nodes[name]; exists {
		return fmt.Errorf("node %s already exists", name)
	}
	r.nodes[name] = Node{Name: name, Weight: weight}
	r.rebuildLocked()
	return nil
}

func (r *Ring) RemoveNode(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[name]; !ok {
		return false
	}
	delete(r.nodes, name)
	r.rebuildLocked()
	return true
}

func (r *Ring) Reset(nodes []Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		if n.Name == "" {
			return errors.New("node name is required")
		}
		if n.Weight < 1 {
			return errors.New("weight must be greater than one")
		}
		if _, exists := r.nodes[n.Name]; exists {
			return fmt.Errorf("node %s already exists", n.Name)
		}
		next[n.Name] = n
	}
	r.nodes = next
	r.rebuildLocked()
	return nil
}

func (r *Ring) Owner(key string) (VirtualNode, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.vnodes) == 0 {
		return VirtualNode{}, false
	}
	h := r.hash(key)
	i := sort.Search(len(r.vnodes), func(i int) bool {
		return r.vnodes[i].Hash >= h
	})
	if i == len(r.vnodes) {
		i = 0
	}
	return r.vnodes[i], true
}

func (r *Ring) Owners(key string, count int) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.vnodes) == 0 || count <= 0 {
		return nil
	}
	if count > len(r.nodes) {
		count = len(r.nodes)
	}
	h := r.hash(key)
	start := sort.Search(len(r.vnodes), func(i int) bool {
		return r.vnodes[i].Hash >= h
	})
	if start == len(r.vnodes) {
		start = 0
	}

	seen := map[string]struct{}{}
	out := make([]Node, 0, count)
	for offset := 0; offset < len(r.vnodes) && len(out) < count; offset++ {
		v := r.vnodes[(start+offset)%len(r.vnodes)]
		if _, ok := seen[v.Node]; ok {
			continue
		}
		seen[v.Node] = struct{}{}
		out = append(out, r.nodes[v.Node])
	}
	return out
}

func (r *Ring) Snapshot() ([]Node, []VirtualNode) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nodes := make([]Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	vnodes := append([]VirtualNode(nil), r.vnodes...)
	return nodes, vnodes
}

func (r *Ring) Hash(key string) uint64 { return r.Hash(key) }

func (r *Ring) rebuildLocked() {
	vnodes := make([]VirtualNode, 0)
	for _, n := range r.nodes {
		count := r.replicasPerUnit * n.Weight
		for replica := 0; replica < count; replica++ {
			token := fmt.Sprintf("%s#%d", n.Name, replica)
			vnodes = append(vnodes, VirtualNode{
				Hash:    r.hash(token),
				Node:    n.Name,
				Replica: replica,
			})
		}
	}
	sort.Slice(vnodes, func(i, j int) bool {
		if vnodes[i].Hash != vnodes[j].Hash {
			return vnodes[i].Hash < vnodes[j].Hash
		}
		if vnodes[i].Node != vnodes[j].Node {
			return vnodes[i].Node < vnodes[j].Node
		}
		return vnodes[i].Replica < vnodes[j].Replica
	})
	r.vnodes = vnodes
}
