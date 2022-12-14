package session

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Store interface {
	Insert(context.Context, *Session) (int64, error)
	GetOne(context.Context, string) (*Session, error)
	GetUsername(context.Context, string) string
	Delete(context.Context, string) error
}

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) Insert(ctx context.Context, d *Session) (int64, error) {
	var id int64
	rows, err := s.db.NamedQuery(`
		INSERT INTO sessions (
			user_id,
			token_hash
		) VALUES (
			:user_id,
			:token_hash
		) RETURNING id`, d)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
	}
	return id, nil
}

func (s *SQLStore) GetOne(ctx context.Context, tokenHash string) (
	*Session,
	error,
) {
	var sessions []*Session
	err := s.db.SelectContext(
		ctx,
		&sessions,
		`SELECT * FROM sessions WHERE token_hash=$1`,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no user exists with this username")
	}
	return sessions[0], nil
}

func (s *SQLStore) GetUsername(ctx context.Context, tokenHash string) string {
	type UserSession struct {
		Username string
	}
	var userSessions []*UserSession
	err := s.db.SelectContext(
		ctx,
		&userSessions,
		`SELECT users.username
		FROM sessions JOIN users ON sessions.user_id = users.id
		WHERE token_hash=$1`,
		tokenHash,
	)
	if err != nil {
		panic(err)
	}
	if len(userSessions) == 0 {
		return ""
	}
	return userSessions[0].Username
}

func (s *SQLStore) Delete(ctx context.Context, tokenHash string) error {
	_, err := s.db.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
