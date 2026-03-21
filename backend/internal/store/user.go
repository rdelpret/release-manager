package store

import (
	"context"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func (s *Store) UpsertUser(ctx context.Context, email, name string, avatarURL *string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, avatar_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET name = $2, avatar_url = $3
		RETURNING id, email, name, avatar_url, created_at
	`, email, name, avatarURL).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, avatar_url, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, avatar_url, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
