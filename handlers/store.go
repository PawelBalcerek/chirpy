package handlers

import (
	"context"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type ChirpStore interface {
	CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error)
	GetChirp(ctx context.Context, id uuid.UUID) (database.Chirp, error)
	GetChirps(ctx context.Context, authorID uuid.UUID) ([]database.Chirp, error)
	DeleteChirp(ctx context.Context, id uuid.UUID) error
}

type UserStore interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUser(ctx context.Context, email string) (database.User, error)
	UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	MakeUserChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error)
	DeleteUsers(ctx context.Context) error
}

type TokenStore interface {
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (database.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}
