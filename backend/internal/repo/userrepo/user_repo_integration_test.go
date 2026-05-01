package userrepo

import (
	"context"
	"testing"

	"atigacbt/backend/internal/model"
	"atigacbt/backend/internal/testutil/pgtest"
)

func TestUserRepo_CreateLookupAndUpdateFlow(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	userID, err := repo.Create(ctx, model.User{
		Username:     "alice",
		PasswordHash: "hash-1",
		Role:         "teacher",
		Name:         "Alice",
		Email:        "alice@example.com",
		GoogleID:     "google-1",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	byUsername, ok, err := repo.GetByUsername(ctx, "alice")
	if err != nil || !ok {
		t.Fatalf("GetByUsername error=%v ok=%v", err, ok)
	}
	if byUsername.ID != userID || byUsername.Email != "alice@example.com" || byUsername.GoogleID != "google-1" {
		t.Fatalf("unexpected user by username: %+v", byUsername)
	}

	byID, ok, err := repo.GetByID(ctx, userID)
	if err != nil || !ok {
		t.Fatalf("GetByID error=%v ok=%v", err, ok)
	}
	if byID.Username != "alice" || byID.Role != "teacher" {
		t.Fatalf("unexpected user by id: %+v", byID)
	}

	byEmail, ok, err := repo.GetByEmail(ctx, "alice@example.com")
	if err != nil || !ok {
		t.Fatalf("GetByEmail error=%v ok=%v", err, ok)
	}
	if byEmail.ID != userID {
		t.Fatalf("expected same user id from email lookup, got %+v", byEmail)
	}

	byGoogleID, ok, err := repo.GetByGoogleID(ctx, "google-1")
	if err != nil || !ok {
		t.Fatalf("GetByGoogleID error=%v ok=%v", err, ok)
	}
	if byGoogleID.ID != userID {
		t.Fatalf("expected same user id from google lookup, got %+v", byGoogleID)
	}

	if err := repo.UpdateProfile(ctx, userID, "Alice Updated", "alice.updated@example.com"); err != nil {
		t.Fatalf("UpdateProfile error: %v", err)
	}
	if err := repo.UpdatePassword(ctx, userID, "hash-2"); err != nil {
		t.Fatalf("UpdatePassword error: %v", err)
	}
	if err := repo.UpdatePhoto(ctx, userID, "/uploads/alice.jpg"); err != nil {
		t.Fatalf("UpdatePhoto error: %v", err)
	}
	if err := repo.UpdateRole(ctx, userID, "admin"); err != nil {
		t.Fatalf("UpdateRole error: %v", err)
	}
	if err := repo.UpdateGoogleID(ctx, userID, "google-2"); err != nil {
		t.Fatalf("UpdateGoogleID error: %v", err)
	}

	updated, ok, err := repo.GetByID(ctx, userID)
	if err != nil || !ok {
		t.Fatalf("GetByID updated error=%v ok=%v", err, ok)
	}
	if updated.Name != "Alice Updated" || updated.Email != "alice.updated@example.com" {
		t.Fatalf("unexpected updated profile: %+v", updated)
	}
	if updated.PasswordHash != "hash-2" || updated.PhotoURL != "/uploads/alice.jpg" {
		t.Fatalf("unexpected updated password/photo: %+v", updated)
	}
	if updated.Role != "admin" || updated.GoogleID != "google-2" {
		t.Fatalf("unexpected updated role/google id: %+v", updated)
	}
}

func TestUserRepo_MissingLookupsReturnFalse(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	if _, ok, err := repo.GetByUsername(ctx, "missing"); err != nil || ok {
		t.Fatalf("GetByUsername missing err=%v ok=%v", err, ok)
	}
	if _, ok, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000"); err != nil || ok {
		t.Fatalf("GetByID missing err=%v ok=%v", err, ok)
	}
	if _, ok, err := repo.GetByEmail(ctx, "missing@example.com"); err != nil || ok {
		t.Fatalf("GetByEmail missing err=%v ok=%v", err, ok)
	}
	if _, ok, err := repo.GetByGoogleID(ctx, "missing-google"); err != nil || ok {
		t.Fatalf("GetByGoogleID missing err=%v ok=%v", err, ok)
	}
}
