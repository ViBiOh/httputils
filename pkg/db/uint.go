package db

import (
	"database/sql/driver"
	"strconv"
	"strings"
)

// WhereInUint wrapper for assigning `IN ($n)` in WHERE clause
type WhereInUint []uint

// Value implements the driver.Valuer interface.
func (a WhereInUint) Value() (driver.Value, error) {
	ints := make([]string, len(a))
	for i, v := range a {
		ints[i] = strconv.FormatUint(uint64(v), 10)
	}

	return strings.Join(ints, ","), nil
}
