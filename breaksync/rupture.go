package breaksync

// Rupture in a break/sync algorithm
type Rupture struct {
	name    string
	first   bool
	last    bool
	extract func(string) string
}

// NewRupture creates and initialize Rupture
func NewRupture(name string, extract func(string) string) *Rupture {
	return &Rupture{first: false, last: true, name: name, extract: extract}
}

// RuptureExtractSimple is a basic extracter that return input
func RuptureExtractSimple(a string) string {
	return a
}

func (a *Rupture) compute(current, next string, force bool) bool {
	a.first = a.last

	if force {
		a.last = true
	} else {
		a.last = a.extract(current) != a.extract(next)
	}

	return a.last
}