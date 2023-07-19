package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token is the type for auth tokens
type Token struct {
	PlainText string    `json:"token"`
	UserId    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken generates a token that lasts for ttl, and returns it
func GenerateToken(userId int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserId: int64(userId),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `delete from tokens where user_id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, u.Id)
	if err != nil {
		return err
	}

	stmt = `
	insert into tokens (user_id, name, email, token_hash, created_at, updated_at, expiry_date)
	values (?, ?, ?, ?, ?, ?, ?)`

	_, err = m.DB.ExecContext(ctx, stmt,
		u.Id,
		u.LastName,
		u.Email,
		t.Hash,
		time.Now(),
		time.Now(),
		t.Expiry,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetUserByToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	query := `
	select users.id, users.first_name, users.last_name, users.email 
	from users
	inner join tokens
	on users.id = tokens.user_id
	where tokens.token_hash = ? and tokens.expiry_date > ?`

	err := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now()).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
