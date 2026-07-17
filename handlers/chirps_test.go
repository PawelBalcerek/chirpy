package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type fakeChirpStore struct {
	chirps []database.Chirp
}

func (f *fakeChirpStore) CreateChirp(_ context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
	chirp := database.Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      arg.Body,
		UserID:    arg.UserID,
	}
	f.chirps = append(f.chirps, chirp)
	return chirp, nil
}

func (f *fakeChirpStore) GetChirp(_ context.Context, id uuid.UUID) (database.Chirp, error) {
	for _, c := range f.chirps {
		if c.ID == id {
			return c, nil
		}
	}
	return database.Chirp{}, sql.ErrNoRows
}

func (f *fakeChirpStore) GetChirps(_ context.Context, _ uuid.UUID) ([]database.Chirp, error) {
	return f.chirps, nil
}

func (f *fakeChirpStore) DeleteChirp(_ context.Context, id uuid.UUID) error {
	for i, c := range f.chirps {
		if c.ID == id {
			f.chirps = append(f.chirps[:i], f.chirps[i+1:]...)
			return nil
		}
	}
	return nil
}
