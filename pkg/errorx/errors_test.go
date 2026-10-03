package errorx_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hiendaovinh/toolkit/pkg/errorx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
)

func TestIsNoRows(t *testing.T) {
	assert.True(t, errorx.IsNoRows(sql.ErrNoRows))
	assert.True(t, errorx.IsNoRows(pgx.ErrNoRows))
	assert.True(t, errorx.IsNoRows(fmt.Errorf("get user: %w", pgx.ErrNoRows)))
	assert.False(t, errorx.IsNoRows(nil))
	assert.False(t, errorx.IsNoRows(errors.New("boom")))
}

func TestIsDuplicated(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", Detail: "Key (email)=(a@b.c) already exists."}

	v, ok := errorx.IsDuplicated(pgErr)
	assert.True(t, ok)
	assert.Equal(t, pgErr, v)

	v, ok = errorx.IsDuplicated(fmt.Errorf("insert user: %w", pgErr))
	assert.True(t, ok)
	assert.Equal(t, pgErr, v)

	_, ok = errorx.IsDuplicated(&pgconn.PgError{Code: "23503"})
	assert.False(t, ok)

	_, ok = errorx.IsDuplicated(nil)
	assert.False(t, ok)

	_, ok = errorx.IsDuplicated(errors.New("boom"))
	assert.False(t, ok)
}

func TestWrap(t *testing.T) {
	cause := errors.New("boom")
	err := errorx.Wrap(cause, errorx.Invalid)
	assert.True(t, err.Of(errorx.Invalid))
	assert.Equal(t, "boom", err.Error())
	assert.Equal(t, "invalid-request", err.Code())
	assert.Equal(t, http.StatusBadRequest, err.Status())
	assert.ErrorIs(t, err, cause)

	err = errorx.Wrap(fmt.Errorf("get user: %w", pgx.ErrNoRows), errorx.Service)
	assert.True(t, err.Of(errorx.NotExist))
	assert.Equal(t, http.StatusNotFound, err.Status())

	err = errorx.Wrap(fmt.Errorf("check: %w", ladon.ErrRequestDenied), errorx.Service)
	assert.True(t, err.Of(errorx.Authz))

	dup := &pgconn.PgError{Code: "23505", Detail: "Key (email)=(a@b.c) already exists."}
	err = errorx.Wrap(fmt.Errorf("insert user: %w", dup), errorx.Service)
	assert.True(t, err.Of(errorx.Exist))
	assert.Equal(t, dup.Detail, err.Error())

	err = errorx.Wrap(fmt.Errorf("insert user: %w", &pgconn.PgError{Code: "23503"}), errorx.Service)
	assert.True(t, err.Of(errorx.Database))
}

func TestMaskErrorMessage(t *testing.T) {
	assert.Equal(t, "unexpected error occurred", errorx.MaskErrorMessage(errors.New("boom")))
	assert.Equal(t, "unable to process", errorx.MaskErrorMessage(errorx.Wrap(errors.New("boom"), errorx.Service)))
	assert.Equal(t, "unable to process", errorx.MaskErrorMessage(errorx.Wrap(&pgconn.PgError{Code: "23503"}, errorx.Service)))
	assert.Equal(t, "boom", errorx.MaskErrorMessage(fmt.Errorf("handler: %w", errorx.Wrap(errors.New("boom"), errorx.Invalid))))
}
