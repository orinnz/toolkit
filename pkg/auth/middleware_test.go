package auth_test

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hiendaovinh/toolkit/pkg/auth"
	"github.com/hiendaovinh/toolkit/pkg/jwtx"
	"github.com/stretchr/testify/assert"
)

func claims(subject string) *jwtx.JWTClaims {
	return &jwtx.JWTClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: subject}}
}

func TestResolveEmptyContext(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", auth.ResolveJWT(ctx))
	assert.Equal(t, "", auth.ResolveSubject(ctx))

	_, ok := auth.ResolveSubjectUUID(ctx)
	assert.False(t, ok)

	_, err := auth.ResolveValidSubject(ctx)
	assert.ErrorIs(t, err, auth.ErrInvalidSession)

	_, err = auth.ResolveValidSubjectUUID(ctx)
	assert.ErrorIs(t, err, auth.ErrInvalidSession)
}

func TestWithAuthJWT(t *testing.T) {
	ctx := auth.WithAuthJWT(context.Background(), "token")
	assert.Equal(t, "token", auth.ResolveJWT(ctx))
}

func TestWithAuthClaimsUUIDSubject(t *testing.T) {
	id := uuid.New()
	ctx := auth.WithAuthClaims(context.Background(), claims(id.String()))

	assert.Equal(t, id.String(), auth.ResolveSubject(ctx))

	got, err := auth.ResolveValidSubjectUUID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, id, got)
}

func TestWithAuthClaimsNonUUIDSubject(t *testing.T) {
	ctx := auth.WithAuthClaims(context.Background(), claims("service-account"))

	sub, err := auth.ResolveValidSubject(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "service-account", sub)

	_, err = auth.ResolveValidSubjectUUID(ctx)
	assert.ErrorIs(t, err, auth.ErrInvalidSession)
}

func TestWithAuthClaimsReplacesSubjectUUID(t *testing.T) {
	ctx := auth.WithAuthClaims(context.Background(), claims(uuid.New().String()))
	ctx = auth.WithAuthClaims(ctx, claims("service-account"))

	assert.Equal(t, "service-account", auth.ResolveSubject(ctx))

	_, ok := auth.ResolveSubjectUUID(ctx)
	assert.False(t, ok, "subject UUID from the previous claims must not leak through")
}
