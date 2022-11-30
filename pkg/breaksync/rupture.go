package breaksync

import "bytes"

var RuptureIdentity = func(a []byte) []byte {
	return []byte(a)
}

type Rupture struct {
	extract func([]byte) []byte
	name    string
	last    bool
	first   bool
}

func NewRupture(name string, extract func([]byte) []byte) *Rupture {
	return &Rupture{
		first:   false,
		last:    true,
		name:    name,
		extract: extract,
	}
}

func (a *Rupture) compute(current, next []byte, force bool) bool {
	a.first = a.last

	if force {
		a.last = true
	} else {
		a.last = !bytes.Equal(a.extract(current), a.extract(next))
	}

	return a.last
}
