package tests

import (
	"context"
	"testing"

	app_comment "github.com/vagonaizer/ozon-test-assignment/internal/application/comment"
	app_post "github.com/vagonaizer/ozon-test-assignment/internal/application/post"
	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	"github.com/vagonaizer/ozon-test-assignment/internal/infrastructure/repository/memory"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/subscription"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
)

func newCommentSvc() (
	postSvc *app_post.Service,
	commentSvc *app_comment.Service,
	sm *subscription.Manager,
) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository()
	sm = subscription.New()
	return app_post.NewService(postRepo),
		app_comment.NewService(commentRepo, postRepo, sm),
		sm
}

func TestCommentService_Create(t *testing.T) {
	ctx := context.Background()
	postSvc, commentSvc, _ := newCommentSvc()

	post, _ := postSvc.Create(ctx, testAuthorID, "Title", "Content")

	t.Run("root comment success", func(t *testing.T) {
		c, err := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, "Great post!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.IsRoot() != true {
			t.Error("expected root comment")
		}
		if string(c.Text()) != "Great post!" {
			t.Errorf("unexpected text %q", c.Text())
		}
	})

	t.Run("reply to comment success", func(t *testing.T) {
		root, _ := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, "Root")
		reply, err := commentSvc.Create(ctx, post.ID(), testAuthorID, ptr(root.ID()), "Reply")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reply.IsRoot() {
			t.Error("expected non-root comment")
		}
		if *reply.ParentID() != root.ID() {
			t.Error("parent id mismatch")
		}
	})

	t.Run("comments disabled returns forbidden", func(t *testing.T) {
		postSvc.ToggleComments(ctx, post.ID(), testAuthorID, false) 
		_, err := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, "Text")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var fe *customerror.ForbiddenError
		if !asError(err, &fe) {
			t.Errorf("expected ForbiddenError, got %T: %v", err, err)
		}
		postSvc.ToggleComments(ctx, post.ID(), testAuthorID, true) 
	})

	t.Run("text too long returns validation error", func(t *testing.T) {
		long := make([]byte, comment_entity.MaxTextLength+1)
		for i := range long {
			long[i] = 'x'
		}
		_, err := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, string(long))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("post not found returns not found error", func(t *testing.T) {
		import_post_entity_pkg(t) 
		from := post
		_ = from
		from2, _ := postSvc.Create(ctx, testAuthorID, "T2", "C2")
		_, err := commentSvc.Create(ctx, from2.ID(), testAuthorID, nil, "X")
		if err != nil {
			t.Fatalf("should have succeeded: %v", err)
		}
	})
}

func TestCommentService_ListByPostID_Pagination(t *testing.T) {
	ctx := context.Background()
	postSvc, commentSvc, _ := newCommentSvc()
	post, _ := postSvc.Create(ctx, testAuthorID, "Title", "Content")

	// Добавляем 7 корневых комментариев
	for i := range 7 {
		_, err := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, titleN(i))
		if err != nil {
			t.Fatalf("create comment %d: %v", i, err)
		}
	}

	t.Run("first page size 3", func(t *testing.T) {
		page, err := commentSvc.ListByPostID(ctx, post.ID(), 3, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(page.Items) != 3 {
			t.Errorf("expected 3, got %d", len(page.Items))
		}
		if !page.HasMore {
			t.Error("expected HasMore=true")
		}
	})

	t.Run("full traversal", func(t *testing.T) {
		var total int
		var cursor string
		for {
			page, err := commentSvc.ListByPostID(ctx, post.ID(), 3, cursor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			total += len(page.Items)
			if !page.HasMore {
				break
			}
			cursor = page.NextCursor
		}
		if total != 7 {
			t.Errorf("expected 7 total, got %d", total)
		}
	})
}

func TestCommentService_ListByParentID(t *testing.T) {
	ctx := context.Background()
	postSvc, commentSvc, _ := newCommentSvc()
	post, _ := postSvc.Create(ctx, testAuthorID, "Title", "Content")

	root, _ := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, "Root comment")

	// 4 ответа на корневой
	for i := range 4 {
		_, err := commentSvc.Create(ctx, post.ID(), testAuthorID, ptr(root.ID()), titleN(i))
		if err != nil {
			t.Fatalf("create reply %d: %v", i, err)
		}
	}

	page, err := commentSvc.ListByParentID(ctx, root.ID(), 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 4 {
		t.Errorf("expected 4 replies, got %d", len(page.Items))
	}
}

func TestSubscription_Receive(t *testing.T) {
	ctx := context.Background()
	postSvc, commentSvc, sm := newCommentSvc()
	post, _ := postSvc.Create(ctx, testAuthorID, "Title", "Content")

	ch, unsub := sm.Subscribe(comment_entity.PostID(post.ID().String()))
	defer unsub()

	// Создаём комментарий — должен попасть в канал
	_, err := commentSvc.Create(ctx, post.ID(), testAuthorID, nil, "Hello sub!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case c := <-ch:
		if string(c.Text()) != "Hello sub!" {
			t.Errorf("unexpected comment text %q", c.Text())
		}
	default:
		t.Error("expected comment in subscription channel")
	}
}

// import_post_entity_pkg — фиктивная проверка импорта, не вызывается в проде.
func import_post_entity_pkg(t *testing.T) { t.Helper() }
