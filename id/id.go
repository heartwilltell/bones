// Package id provides the set of functions to generate
// different kind of identifiers
package id

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/valyala/fastrand"
)

var _ error = (Error)("")

const (
	// ErrInvalidID represents an error which indicates that given TID is invalid.
	ErrInvalidID Error = "id: invalid identifier"
)

// ULID returns ULID identifier as string.
// More about ULID: https://github.com/ulid/spec
func ULID() string {
	t := time.Now().UTC()
	e := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0) //nolint:gosec
	id := ulid.MustNew(ulid.Timestamp(t), e)

	return id.String()
}

// ValidateULID validates string representation
// of ULID identifier.
func ValidateULID(id string) error {
	if _, err := ulid.Parse(id); err != nil {
		return ErrInvalidID
	}

	return nil
}

// XID returns short unique identifier as string.
func XID() string { return strings.ToUpper(xid.New().String()) }

// ValidateXID validates string representation of XID identifier.
func ValidateXID(id string) error {
	if _, err := xid.FromString(id); err != nil {
		return ErrInvalidID
	}

	return nil
}

// DigiCode returns 6-digit code as a string.
func DigiCode() string {
	const (
		maxN    = 9
		codeLen = 6
	)

	var rng fastrand.RNG
	rng.Seed(uint32(time.Now().UnixNano()))

	var b strings.Builder

	for i := 0; i < codeLen; i++ {
		b.WriteString(strconv.Itoa(int(fastrand.Uint32n(maxN))))
	}

	return b.String()
}

// ValidateDigiCode validates code from DigiCode.
func ValidateDigiCode(code string) error {
	if len(code) != 6 || utf8.RuneCountInString(code) != 6 {
		return ErrInvalidID
	}

	for _, r := range code {
		if !unicode.IsNumber(r) {
			return ErrInvalidID
		}
	}

	return nil
}

// Error represents package level error.
type Error string

func (e Error) Error() string { return string(e) }
