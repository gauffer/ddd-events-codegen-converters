package model

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"
	"strings"
)

type Phone string

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

func NewPhone(number string) Phone {
	return normalizePhone(number)
}

func normalizePhone(phone string) Phone {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if strings.HasPrefix(phone, "8") {
		phone = "+7" + phone[1:]
	}

	return Phone(phone)
}

func (p Phone) IsValid() bool {
	return phoneRegex.MatchString(string(p))
}

func (p Phone) Masked() string {
	if len(p) < 8 {
		return string(p)
	}

	return string(p[:4]) + "****" + string(p[len(p)-4:])
}

type ProofKeyMethod string

const (
	S256 ProofKeyMethod = "S256"
)

func (m ProofKeyMethod) IsValid() bool {
	return m == S256
}

type ProofKey struct {
	Value  string
	Method ProofKeyMethod
}

func NewProofKey(value string, method ProofKeyMethod) (*ProofKey, error) {
	if !method.IsValid() {
		return nil, ErrInvalidCodeChallengeMethod
	}

	return &ProofKey{
		Value:  value,
		Method: method,
	}, nil
}

func (p ProofKey) Verify(source string) bool {
	if len(source) < 43 || len(source) > 128 {
		return false
	}

	switch p.Method {
	case S256:
		hash := sha256.Sum256([]byte(source))
		expected := base64.RawURLEncoding.EncodeToString(hash[:])

		if subtle.ConstantTimeCompare([]byte(expected), []byte(p.Value)) == 1 {
			return true
		}

	default:
		return false
	}

	return false
}
