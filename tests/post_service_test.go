package tests

import (
	"context"
	"testing"

	app_post "github.com/vagonaizer/ozon-test-assignment/internal/application/post"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/infrastructure/repository/memory"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
)

func TestPostService_Create(t *testing.T) {
	svc := app_post.NewService(memory.NewPostRepository())
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		post, err := svc.Create(ctx, testAuthorID, "Hello World", "Content here")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.Title().String() != "Hello World" {
			t.Errorf("expected title 'Hello World', got %q", post.Title().String())
		}
		if !post.CommentsEnabled() {
			t.Error("expected comments to be enabled by default")
		}
	})

	t.Run("empty title returns validation error", func(t *testing.T) {
		_, err := svc.Create(ctx, testAuthorID, "", "Content")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ve *customerror.ValidationError
		if !asError(err, &ve) {
			t.Errorf("expected ValidationError, got %T: %v", err, err)
		}
	})

	t.Run("empty content returns validation error", func(t *testing.T) {
		_, err := svc.Create(ctx, testAuthorID, "Title", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid author id returns validation error", func(t *testing.T) {
		_, err := svc.Create(ctx, "not-a-ulid", "Title", "Content")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostService_GetByID(t *testing.T) {
	repo := memory.NewPostRepository()
	svc := app_post.NewService(repo)
	ctx := context.Background()

	post, _ := svc.Create(ctx, testAuthorID, "Title", "Content")

	t.Run("found", func(t *testing.T) {
		got, err := svc.GetByID(ctx, post.ID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID().String() != post.ID().String() {
			t.Errorf("id mismatch")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetByID(ctx, post_entity.NewPostID())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var nfe *customerror.NotFoundError
		if !asError(err, &nfe) {
			t.Errorf("expected NotFoundError, got %T: %v", err, err)
		}
	})
}

func TestPostService_List_Pagination(t *testing.T) {
	repo := memory.NewPostRepository()
	svc := app_post.NewService(repo)
	ctx := context.Background()

	// Создаём 5 постов
	for i := range 5 {
		_, err := svc.Create(ctx, testAuthorID, titleN(i), "Content")
		if err != nil {
			t.Fatalf("create post %d: %v", i, err)
		}
	}

	t.Run("first page", func(t *testing.T) {
		page, err := svc.List(ctx, 2, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(page.Items))
		}
		if !page.HasMore {
			t.Error("expected HasMore=true")
		}
		if page.NextCursor == "" {
			t.Error("expected non-empty NextCursor")
		}
	})

	t.Run("full traversal", func(t *testing.T) {
		var total int
		var cursor string
		for {
			page, err := svc.List(ctx, 2, cursor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			total += len(page.Items)
			if !page.HasMore {
				break
			}
			cursor = page.NextCursor
		}
		if total != 5 {
			t.Errorf("expected 5 total posts, got %d", total)
		}
	})
}

func TestPostService_ToggleComments(t *testing.T) {
	repo := memory.NewPostRepository()
	svc := app_post.NewService(repo)
	ctx := context.Background()

	post, _ := svc.Create(ctx, testAuthorID, "Title", "Content")

	t.Run("author can disable comments", func(t *testing.T) {
		updated, err := svc.ToggleComments(ctx, post.ID(), testAuthorID, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.CommentsEnabled() {
			t.Error("expected comments to be disabled")
		}
	})

	t.Run("non-author gets forbidden error", func(t *testing.T) {
		otherID := "01JQHX00000000000000000002"
		_, err := svc.ToggleComments(ctx, post.ID(), otherID, true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var fe *customerror.ForbiddenError
		if !asError(err, &fe) {
			t.Errorf("expected ForbiddenError, got %T: %v", err, err)
		}
	})
}
