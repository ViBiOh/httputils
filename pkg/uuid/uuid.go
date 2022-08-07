package uuid

import (
	"crypto/rand"
	"fmt"
)

// New generate a uuid
func New() (string, error) {
	raw := make([]byte, 16)
	_, err := rand.Read(raw)
	if err != nil {
		return "", fmt.Errorf("read random: %s", err)
	}

	raw[6] = raw[6]&^0xf0 | 0x40 // set version to 4 (random uuid)
	raw[8] = raw[8]&^0xc0 | 0x80 // set to IETF variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:]), nil
}
