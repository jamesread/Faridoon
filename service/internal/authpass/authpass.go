package authpass

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"github.com/alexedwards/argon2id"
	"golang.org/x/crypto/bcrypt"
)

// Verify checks plain against stored hash (argon2id, bcrypt, or legacy sha1).
func Verify(plain, stored string) bool {
	stored = strings.TrimSpace(stored)
	if stored == "" || plain == "" {
		return false
	}
	return verifyStored(plain, stored)
}

func verifyStored(plain, stored string) bool {
	if strings.HasPrefix(stored, "$argon2id$") {
		return verifyArgon2(plain, stored)
	}
	if isBcrypt(stored) {
		return verifyBcrypt(plain, stored)
	}
	return verifyLegacySHA1(plain, stored)
}

func isBcrypt(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") ||
		strings.HasPrefix(stored, "$2b$") ||
		strings.HasPrefix(stored, "$2y$")
}

func verifyArgon2(plain, stored string) bool {
	ok, err := argon2id.ComparePasswordAndHash(plain, stored)
	return err == nil && ok
}

func verifyBcrypt(plain, stored string) bool {
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
}

func verifyLegacySHA1(plain, stored string) bool {
	if len(stored) != 40 {
		return false
	}
	sum := sha1.Sum([]byte(plain))
	return strings.EqualFold(hex.EncodeToString(sum[:]), stored)
}

// NeedsRehash reports whether stored should be upgraded to argon2id.
func NeedsRehash(stored string) bool {
	return !strings.HasPrefix(strings.TrimSpace(stored), "$argon2id$")
}

// Hash returns an argon2id hash of plain.
func Hash(plain string) (string, error) {
	return argon2id.CreateHash(plain, argon2id.DefaultParams)
}
