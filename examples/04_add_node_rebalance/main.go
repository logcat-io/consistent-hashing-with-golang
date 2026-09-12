package main

import (
	"fmt"

	"github.com/logcat/consistent-hashing-go-lab/internal/hashring"
)

func owners(nodes []hashring.Node) map[string]string {
	r := hashring.New(200, nil)
	_ = r.Reset(nodes)
	m := map[string]string{}
	for i := 0; i < 100_000; i++ {
		k := fmt.Sprintf("user:%d", i)
		v, _ := r.Owner(k)
		m[k] = v.Node
	}
	return m
}

func main() {
	a := owners([]hashring.Node{{"a", 1}, {"b", 1}, {"c", 1}})
	b := owners([]hashring.Node{{"a", 1}, {"b", 1}, {"c", 1}, {"d", 1}})
	m := 0
	for k, v := range a {
		if b[k] != v {
			m++
		}
	}
	fmt.Printf("moved %2.f%% (expected around 25%%)", float64(m)/1000)
}
