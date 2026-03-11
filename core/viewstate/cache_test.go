package viewstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFingerprintIsStable(t *testing.T) {
	assert.Equal(t, HashFingerprint("skills", "meta"), HashFingerprint("skills", "meta"))
	assert.NotEqual(t, HashFingerprint("skills", "meta"), HashFingerprint("skills", "config"))
}
