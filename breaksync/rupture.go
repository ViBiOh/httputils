package breaksync

type Rupture struct {
	first   bool
	last    bool
	extract func(string) string
}

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
