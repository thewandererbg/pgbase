package security

import (
	"math/rand"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/oklog/ulid/v2"
)

func GeneratePrefixedULID(prefix string) string {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	u, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		u, _ = ulid.New(ulid.Timestamp(t), entropy)
	}
	return prefix + "_" + u.String()
}

func GenerateULID() string {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	u, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		u, _ = ulid.New(ulid.Timestamp(t), entropy)
	}
	return u.String()
}

func GenerateUUIDv4() string {
	if u, err := uuid.NewV4(); err == nil {
		return u.String()
	}
	u, _ := uuid.NewV4()
	return u.String()
}

func GenerateUUIDv7() string {
	if u, err := uuid.NewV7(); err == nil {
		return u.String()
	}
	u, _ := uuid.NewV7()
	return u.String()
}
