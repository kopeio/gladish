package sets

import "sort"

// Borrowed from k8s

type Empty struct{}

// sets.String is a set of strings, implemented via map[string]struct{} for minimal memory consumption.
type String map[string]Empty

// New creates a String from a list of values.
func NewString(items ...string) String {
	ss := String{}
	ss.Insert(items...)
	return ss
}
func (ss String) Insert(items ...string) {
	for _, s := range items {
		ss[s] = Empty{}
	}
}

// Has returns true if and only if item is contained in the set.
func (s String) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// List returns the contents as a sorted string slice.
func (s String) List() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}
