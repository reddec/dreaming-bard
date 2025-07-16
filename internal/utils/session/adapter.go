package session

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/reddec/dreaming-bard/internal/dbo"
)

func NewDBSession(db *dbo.Queries) *DBSession {
	return &DBSession{db: db}
}

type DBSession struct {
	db *dbo.Queries
}

func (dbs *DBSession) Delete(token string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return dbs.DeleteCtx(ctx, token)
}

func (dbs *DBSession) Find(token string) (b []byte, found bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return dbs.FindCtx(ctx, token)
}

func (dbs *DBSession) Commit(token string, b []byte, expiry time.Time) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return dbs.CommitCtx(ctx, token, b, expiry)
}

func (dbs *DBSession) DeleteCtx(ctx context.Context, token string) (err error) {
	return dbs.db.DeleteSession(ctx, token)
}

func (dbs *DBSession) FindCtx(ctx context.Context, token string) (b []byte, found bool, err error) {
	info, err := dbs.db.FindSession(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if info.ExpiresAt.Before(time.Now()) {
		return nil, false, dbs.DeleteCtx(ctx, token)
	}
	return info.Content, true, nil
}

func (dbs *DBSession) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error) {
	return dbs.db.SetSession(ctx, dbo.SetSessionParams{
		Token:     token,
		Content:   b,
		ExpiresAt: expiry,
	})
}

func (dbs *DBSession) GC(ctx context.Context) error {
	return dbs.db.DeleteSessionExpired(ctx)
}
