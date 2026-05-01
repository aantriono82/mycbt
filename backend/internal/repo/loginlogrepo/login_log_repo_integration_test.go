package loginlogrepo

import (
	"context"
	"testing"
	"time"

	"atigacbt/backend/internal/testutil/pgtest"
)

func TestLoginLogRepo_InsertListAndCleanup(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	userID := "11111111-1111-1111-1111-111111111111"
	if _, err := h.Pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, role, name, email, is_active)
		VALUES ($1::uuid, 'admin-log', 'hash', 'admin', 'Admin Log', 'admin-log@example.com', true)
	`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	oldTime := time.Now().UTC().AddDate(0, 0, -40)
	newTime := time.Now().UTC().Add(-time.Hour)
	if err := repo.Insert(ctx, LoginLog{
		UserID:     &userID,
		Username:   "admin-log",
		Role:       "admin",
		IP:         "10.0.0.1",
		UserAgent:  "UA-1",
		LoggedInAt: oldTime,
	}); err != nil {
		t.Fatalf("Insert old log error: %v", err)
	}
	if err := repo.Insert(ctx, LoginLog{
		UserID:     &userID,
		Username:   "admin-log",
		Role:       "admin",
		IP:         "10.0.0.2",
		UserAgent:  "UA-2",
		LoggedInAt: newTime,
	}); err != nil {
		t.Fatalf("Insert new log error: %v", err)
	}

	items, total, err := repo.List(ctx, ListFilter{Q: "admin", Role: "admin", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 logs, total=%d len=%d", total, len(items))
	}
	if items[0].IP != "10.0.0.2" {
		t.Fatalf("expected newest log first, got %+v", items)
	}

	itemsByIP, totalByIP, err := repo.List(ctx, ListFilter{IP: "10.0.0.1", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List by IP error: %v", err)
	}
	if totalByIP != 1 || len(itemsByIP) != 1 {
		t.Fatalf("expected 1 IP-matched log, total=%d len=%d", totalByIP, len(itemsByIP))
	}

	pruned, err := repo.PruneOlderThan(ctx, 30)
	if err != nil {
		t.Fatalf("PruneOlderThan error: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("expected 1 pruned row, got %d", pruned)
	}

	remaining, totalRemaining, err := repo.List(ctx, ListFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List remaining error: %v", err)
	}
	if totalRemaining != 1 || len(remaining) != 1 {
		t.Fatalf("expected 1 remaining log, total=%d len=%d", totalRemaining, len(remaining))
	}

	if err := repo.DeleteByID(ctx, remaining[0].ID); err != nil {
		t.Fatalf("DeleteByID error: %v", err)
	}
	cleared, err := repo.ClearAll(ctx)
	if err != nil {
		t.Fatalf("ClearAll error: %v", err)
	}
	if cleared != 0 {
		t.Fatalf("expected 0 rows cleared after explicit delete, got %d", cleared)
	}
}
