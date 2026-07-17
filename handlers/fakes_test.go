package handlers_test

import (
	"context"
	"database/sql"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type fakeChirpStore struct {
	CreateChirpFunc func(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error)
	GetChirpFunc    func(ctx context.Context, id uuid.UUID) (database.Chirp, error)
	GetChirpsFunc   func(ctx context.Context, authorID uuid.UUID) ([]database.Chirp, error)
	DeleteChirpFunc func(ctx context.Context, id uuid.UUID) error
}

func (f *fakeChirpStore) CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
	if f.CreateChirpFunc != nil {
		return f.CreateChirpFunc(ctx, arg)
	}
	return database.Chirp{}, nil
}

func (f *fakeChirpStore) GetChirp(ctx context.Context, id uuid.UUID) (database.Chirp, error) {
	if f.GetChirpFunc != nil {
		return f.GetChirpFunc(ctx, id)
	}
	return database.Chirp{}, sql.ErrNoRows
}

func (f *fakeChirpStore) GetChirps(ctx context.Context, authorID uuid.UUID) ([]database.Chirp, error) {
	if f.GetChirpsFunc != nil {
		return f.GetChirpsFunc(ctx, authorID)
	}
	return nil, nil
}

func (f *fakeChirpStore) DeleteChirp(ctx context.Context, id uuid.UUID) error {
	if f.DeleteChirpFunc != nil {
		return f.DeleteChirpFunc(ctx, id)
	}
	return nil
}

type fakeUserStore struct {
	CreateUserFunc        func(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserFunc           func(ctx context.Context, email string) (database.User, error)
	UpdateUserFunc        func(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
	MakeUserChirpyRedFunc func(ctx context.Context, id uuid.UUID) (database.User, error)
	DeleteUsersFunc       func(ctx context.Context) error
}

func (f *fakeUserStore) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	if f.CreateUserFunc != nil {
		return f.CreateUserFunc(ctx, arg)
	}
	return database.User{}, nil
}

func (f *fakeUserStore) GetUser(ctx context.Context, email string) (database.User, error) {
	if f.GetUserFunc != nil {
		return f.GetUserFunc(ctx, email)
	}
	return database.User{}, sql.ErrNoRows
}

func (f *fakeUserStore) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	if f.UpdateUserFunc != nil {
		return f.UpdateUserFunc(ctx, arg)
	}
	return database.User{}, nil
}

func (f *fakeUserStore) MakeUserChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error) {
	if f.MakeUserChirpyRedFunc != nil {
		return f.MakeUserChirpyRedFunc(ctx, id)
	}
	return database.User{}, nil
}

func (f *fakeUserStore) DeleteUsers(ctx context.Context) error {
	if f.DeleteUsersFunc != nil {
		return f.DeleteUsersFunc(ctx)
	}
	return nil
}

type fakeTokenStore struct {
	CreateRefreshTokenFunc func(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
	GetRefreshTokenFunc    func(ctx context.Context, token string) (database.RefreshToken, error)
	RevokeRefreshTokenFunc func(ctx context.Context, token string) error
}

func (f *fakeTokenStore) CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error) {
	if f.CreateRefreshTokenFunc != nil {
		return f.CreateRefreshTokenFunc(ctx, arg)
	}
	return database.RefreshToken{}, nil
}

func (f *fakeTokenStore) GetRefreshToken(ctx context.Context, token string) (database.RefreshToken, error) {
	if f.GetRefreshTokenFunc != nil {
		return f.GetRefreshTokenFunc(ctx, token)
	}
	return database.RefreshToken{}, sql.ErrNoRows
}

func (f *fakeTokenStore) RevokeRefreshToken(ctx context.Context, token string) error {
	if f.RevokeRefreshTokenFunc != nil {
		return f.RevokeRefreshTokenFunc(ctx, token)
	}
	return nil
}
