package inputpart_test

import (
	"context"
	"testing"

	"go.orx.me/apps/butter/internal/repo/inputpart"
	"go.orx.me/apps/butter/internal/repo/inputpart/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newMemoryRepo() inputpart.Repository { return memory.New() }

func runContract(t *testing.T, repo inputpart.Repository) {
	t.Helper()
	ctx := context.Background()

	t.Run("SaveAll_and_Load_text_only", func(t *testing.T) {
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "hello world"}},
		}
		if err := repo.SaveAll(ctx, "inv-text", parts); err != nil {
			t.Fatalf("SaveAll: %v", err)
		}
		got, err := repo.Load(ctx, "inv-text")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Load: got %d parts, want 1", len(got))
		}
		if got[0].GetText() != "hello world" {
			t.Fatalf("text = %q, want %q", got[0].GetText(), "hello world")
		}
	})

	t.Run("SaveAll_and_Load_image_only", func(t *testing.T) {
		imgData := []byte("fake-png-data")
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_InlineData{InlineData: &agentsv1.InlineData{
				MimeType: "image/png",
				Data:     imgData,
			}}},
		}
		if err := repo.SaveAll(ctx, "inv-img", parts); err != nil {
			t.Fatalf("SaveAll: %v", err)
		}
		got, err := repo.Load(ctx, "inv-img")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("Load: got %d parts, want 1", len(got))
		}
		inline := got[0].GetInlineData()
		if inline == nil {
			t.Fatal("expected InlineData, got nil")
		}
		if inline.GetMimeType() != "image/png" {
			t.Fatalf("mime = %q, want image/png", inline.GetMimeType())
		}
		if string(inline.GetData()) != string(imgData) {
			t.Fatal("image data mismatch")
		}
	})

	t.Run("SaveAll_and_Load_mixed_preserves_order", func(t *testing.T) {
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "first"}},
			{Part: &agentsv1.InputPart_InlineData{InlineData: &agentsv1.InlineData{
				MimeType: "image/jpeg",
				Data:     []byte("jpg-bytes"),
			}}},
			{Part: &agentsv1.InputPart_Text{Text: "second"}},
		}
		if err := repo.SaveAll(ctx, "inv-mixed", parts); err != nil {
			t.Fatalf("SaveAll: %v", err)
		}
		got, err := repo.Load(ctx, "inv-mixed")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("Load: got %d parts, want 3", len(got))
		}
		if got[0].GetText() != "first" {
			t.Fatalf("part[0] text = %q, want first", got[0].GetText())
		}
		if got[1].GetInlineData().GetMimeType() != "image/jpeg" {
			t.Fatalf("part[1] mime = %q, want image/jpeg", got[1].GetInlineData().GetMimeType())
		}
		if got[2].GetText() != "second" {
			t.Fatalf("part[2] text = %q, want second", got[2].GetText())
		}
	})

	t.Run("SaveAll_idempotent", func(t *testing.T) {
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "once"}},
		}
		if err := repo.SaveAll(ctx, "inv-idem", parts); err != nil {
			t.Fatalf("first SaveAll: %v", err)
		}
		// Second save with different content should be no-op.
		parts2 := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "twice"}},
			{Part: &agentsv1.InputPart_Text{Text: "thrice"}},
		}
		if err := repo.SaveAll(ctx, "inv-idem", parts2); err != nil {
			t.Fatalf("second SaveAll: %v", err)
		}
		got, err := repo.Load(ctx, "inv-idem")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(got) != 1 || got[0].GetText() != "once" {
			t.Fatalf("idempotency violated: got %d parts with text=%q", len(got), got[0].GetText())
		}
	})

	t.Run("Load_not_found", func(t *testing.T) {
		_, err := repo.Load(ctx, "inv-nonexistent")
		if err != inputpart.ErrNotFound {
			t.Fatalf("Load: want ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete_removes_all_parts", func(t *testing.T) {
		parts := []*agentsv1.InputPart{
			{Part: &agentsv1.InputPart_Text{Text: "del-1"}},
			{Part: &agentsv1.InputPart_Text{Text: "del-2"}},
		}
		if err := repo.SaveAll(ctx, "inv-del", parts); err != nil {
			t.Fatalf("SaveAll: %v", err)
		}
		if err := repo.Delete(ctx, "inv-del"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.Load(ctx, "inv-del")
		if err != inputpart.ErrNotFound {
			t.Fatalf("after Delete: want ErrNotFound, got %v", err)
		}
	})

	t.Run("Delete_nonexistent_is_noop", func(t *testing.T) {
		if err := repo.Delete(ctx, "inv-nope"); err != nil {
			t.Fatalf("Delete nonexistent: %v", err)
		}
	})

	t.Run("SaveAll_large_payload", func(t *testing.T) {
		// Simulate 10 images at ~1 MiB each (within the 10 MiB per-image limit).
		imgData := make([]byte, 1<<20) // 1 MiB
		for i := range imgData {
			imgData[i] = byte(i % 256)
		}
		parts := make([]*agentsv1.InputPart, 10)
		for i := range parts {
			parts[i] = &agentsv1.InputPart{Part: &agentsv1.InputPart_InlineData{InlineData: &agentsv1.InlineData{
				MimeType: "image/webp",
				Data:     imgData,
			}}}
		}
		if err := repo.SaveAll(ctx, "inv-large", parts); err != nil {
			t.Fatalf("SaveAll large: %v", err)
		}
		got, err := repo.Load(ctx, "inv-large")
		if err != nil {
			t.Fatalf("Load large: %v", err)
		}
		if len(got) != 10 {
			t.Fatalf("Load large: got %d parts, want 10", len(got))
		}
		for i, p := range got {
			if len(p.GetInlineData().GetData()) != 1<<20 {
				t.Fatalf("part[%d] size = %d, want %d", i, len(p.GetInlineData().GetData()), 1<<20)
			}
		}
	})
}

func TestMemoryContract(t *testing.T) {
	runContract(t, newMemoryRepo())
}
