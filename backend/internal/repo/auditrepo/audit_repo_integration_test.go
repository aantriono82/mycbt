package auditrepo

import (
	"context"
	"testing"
	"time"

	"atigacbt/backend/internal/testutil/pgtest"
)

func TestAuditRepo_CreateListAndCleanup(t *testing.T) {
	h := pgtest.Setup(t)
	ctx := context.Background()
	repo := New(h.Pool)

	userID := "22222222-2222-2222-2222-222222222222"
	if _, err := h.Pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, role, name, email, is_active)
		VALUES ($1::uuid, 'audit-user', 'hash', 'admin', 'Audit User', 'audit-user@example.com', true)
	`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := repo.Create(ctx, CreateLogInput{
		RequestID:  "req-1",
		UserID:     userID,
		Role:       "admin",
		Method:     "POST",
		Path:       "/api/v1/exams",
		Query:      "draft=1",
		StatusCode: 201,
		IP:         "10.10.0.1",
		UserAgent:  "browser-1",
		Payload:    map[string]any{"title": "Exam 1"},
	}); err != nil {
		t.Fatalf("Create first audit log error: %v", err)
	}
	if err := repo.Create(ctx, CreateLogInput{
		RequestID:  "req-2",
		UserID:     userID,
		Role:       "admin",
		Method:     "GET",
		Path:       "/api/v1/exams",
		Query:      "",
		StatusCode: 200,
		IP:         "10.10.0.2",
		UserAgent:  "browser-2",
		Payload:    map[string]any{"ok": true},
	}); err != nil {
		t.Fatalf("Create second audit log error: %v", err)
	}

	items, total, err := repo.List(ctx, ListFilter{Q: "req-", Role: "admin", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 audit logs, total=%d len=%d", total, len(items))
	}
	if items[0].RequestID != "req-2" {
		t.Fatalf("expected newest audit log first, got %+v", items)
	}
	if items[0].Payload["ok"] != true {
		t.Fatalf("expected payload to be decoded, got %+v", items[0].Payload)
	}

	status := 201
	filtered, totalFiltered, err := repo.List(ctx, ListFilter{
		Method: "POST",
		Path:   "/api/v1/exams",
		Status: &status,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("filtered List error: %v", err)
	}
	if totalFiltered != 1 || len(filtered) != 1 || filtered[0].RequestID != "req-1" {
		t.Fatalf("unexpected filtered result total=%d items=%+v", totalFiltered, filtered)
	}

	if err := repo.DeleteByID(ctx, filtered[0].ID); err != nil {
		t.Fatalf("DeleteByID error: %v", err)
	}

	oldTime := time.Now().UTC().AddDate(0, 0, -40)
	if _, err := h.Pool.Exec(ctx, `
		UPDATE audit_logs SET created_at = $2 WHERE request_id = $1
	`, "req-2", oldTime); err != nil {
		t.Fatalf("age audit log error: %v", err)
	}
	pruned, err := repo.PruneOlderThan(ctx, 30)
	if err != nil {
		t.Fatalf("PruneOlderThan error: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("expected 1 pruned audit log, got %d", pruned)
	}

	cleared, err := repo.ClearAll(ctx)
	if err != nil {
		t.Fatalf("ClearAll error: %v", err)
	}
	if cleared != 0 {
		t.Fatalf("expected 0 rows after prune, got %d", cleared)
	}
}
