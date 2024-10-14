package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func Hash(originalURL string) string {
	timestamp := time.Now().UnixNano()

	dataToHash := fmt.Sprintf("%s%d", originalURL, timestamp)

	hash := sha256.Sum256([]byte(dataToHash))

	encoded := base64.URLEncoding.EncodeToString(hash[:])

	shortened := encoded[:8]

	return shortened
}

func ConvertFromUuidPg(id uuid.UUID) pgtype.UUID {
    return pgtype.UUID{
        Bytes: id,
        Valid: true,
    }
}

func ConvertFromPgUuid(id pgtype.UUID) uuid.UUID {
    return id.Bytes
}