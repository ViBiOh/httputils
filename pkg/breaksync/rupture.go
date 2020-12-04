package breaksync

// Rupture in a break/sync algorithm
type Rupture struct {
	extract func(string) string
	name    string
	last    bool
	first   bool
}

// NewRupture creates and initialize Rupture
func NewRupture(name string, extract func(string) string) *Rupture {
	return &Rupture{
		first:   false,
		last:    true,
		name:    name,
		extract: extract,
	}
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
