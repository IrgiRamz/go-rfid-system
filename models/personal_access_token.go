package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/yimm/rfid-api/config"
)

type PersonalAccessToken struct {
	ID            int64           `json:"id"`
	TokenableType string          `json:"tokenable_type"`
	TokenableID   int64           `json:"tokenable_id"`
	Name          string          `json:"name"`
	Token         string          `json:"-"`
	Abilities     sql.NullString  `json:"-"`
	LastUsedAt    sql.NullTime    `json:"-"`
	ExpiresAt     sql.NullTime    `json:"-"`
	CreatedAt     time.Time       `json:"-"`
	UpdatedAt     time.Time       `json:"-"`
}

func hashToken(plainToken string) string {
	hash := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(hash[:])
}

func GeneratePlainToken() (string, error) {
	bytes := make([]byte, 40)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CreateToken(userID int64, name string, expiresAt *time.Time) (int64, string, error) {
	plainToken, err := GeneratePlainToken()
	if err != nil {
		return 0, "", err
	}

	tokenHash := hashToken(plainToken)
	abilities := `"[*]"`

	var result sql.Result
	if expiresAt != nil {
		result, err = config.DB.Exec(`
			INSERT INTO personal_access_tokens
			(tokenable_type, tokenable_id, name, token, abilities, expires_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		`, "App\\Models\\User", userID, name, tokenHash, abilities, expiresAt)
	} else {
		result, err = config.DB.Exec(`
			INSERT INTO personal_access_tokens
			(tokenable_type, tokenable_id, name, token, abilities, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		`, "App\\Models\\User", userID, name, tokenHash, abilities)
	}
	if err != nil {
		return 0, "", err
	}

	tokenID, err := result.LastInsertId()
	if err != nil {
		return 0, "", err
	}

	return tokenID, plainToken, nil
}

func ValidateToken(tokenID int64, plainToken string) (*PersonalAccessToken, error) {
	tokenHash := hashToken(plainToken)

	token := &PersonalAccessToken{}
	query := `
		SELECT id, tokenable_type, tokenable_id, name, abilities, last_used_at, expires_at, created_at, updated_at
		FROM personal_access_tokens
		WHERE id = ? AND token = ?
	`
	err := config.DB.QueryRow(query, tokenID, tokenHash).Scan(
		&token.ID, &token.TokenableType, &token.TokenableID, &token.Name,
		&token.Abilities, &token.LastUsedAt, &token.ExpiresAt,
		&token.CreatedAt, &token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if token.ExpiresAt.Valid && token.ExpiresAt.Time.Before(time.Now()) {
		return nil, sql.ErrNoRows
	}

	config.DB.Exec(`UPDATE personal_access_tokens SET last_used_at = NOW() WHERE id = ?`, tokenID)

	return token, nil
}

func DeleteToken(tokenID int64) error {
	_, err := config.DB.Exec(`DELETE FROM personal_access_tokens WHERE id = ?`, tokenID)
	return err
}

func DeleteUserTokens(userID int64) error {
	_, err := config.DB.Exec(`
		DELETE FROM personal_access_tokens
		WHERE tokenable_type = 'App\\Models\\User' AND tokenable_id = ?
	`, userID)
	return err
}