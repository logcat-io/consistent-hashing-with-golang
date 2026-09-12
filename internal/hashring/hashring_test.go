package hashring

import "testing"

func TestOwnerIsStable(t *testing.T) {
	r := New(100, SHA256Hash64)
	if err := r.Reset([]Node{{"a", 1}, {"b", 1}, {"c", 1}}); err != nil {
		t.Fatal(err)
	}
	first, ok := r.Owner("user:1001")
	if !ok {
		t.Fatal("expected owner")
	}
	for i := 0; i < 100; i++ {
		got, _ := r.Owner("user:1001")
		if got.Node != first.Node || got.Hash != first.Hash {
			t.Fatalf("owner changed: first=%+v got=%+v", first, got)
		}
	}
}

func TestOwnersAreDistinctPhysicalNodes(t *testing.T) {
	r := New(100, SHA256Hash64)
	if err := r.Reset([]Node{{"a", 1}, {"b", 1}, {"c", 1}, {"d", 1}}); err != nil {
		t.Fatal(err)
	}
	owners := r.Owners("user:1001", 3)
	if len(owners) != 3 {
		t.Fatalf("want 3, got %d", len(owners))
	}
	seen := map[string]bool{}
	for _, n := range owners {
		if seen[n.Name] {
			t.Fatalf("duplicate physical node: %s", n.Name)
		}
		seen[n.Name] = true
	}
}

func TestWrapAround(t *testing.T) {
	hash := func(s string) uint64 {
		switch s {
		case "a#0":
			return 10
		case "b#0":
			return 20
		case "key":
			return 99
		default:
			return 0
		}
	}
	r := New(1, hash)
	if err := r.Reset([]Node{{"a", 1}, {"b", 1}}); err != nil {
		t.Fatal(err)
	}
	v, ok := r.Owner("key")
	if !ok || v.Node != "a" {
		t.Fatalf("want wrap to a, got %+v ok=%v", v, ok)
	}
}

func TestCollisionOrderingIsDeterministic(t *testing.T) {
	hash := func(string) uint64 { return 7 }
	r := New(1, hash)
	if err := r.Reset([]Node{{"b", 1}, {"a", 1}}); err != nil {
		t.Fatal(err)
	}
	v, ok := r.Owner("key")
	if !ok {
		t.Fatal("expected owner")
	}
	if v.Node != "a" {
		t.Fatalf("tie-break should choose a first, got %s", v.Node)
	}
}
