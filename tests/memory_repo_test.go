package tests

import (
	"context"
	"testing"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/infrastructure/repository/memory"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
)

// ─── PostRepository ────────────────────────────────────────────────────────

func TestMemoryPostRepo_CreateAndGet(t *testing.T) {
	repo := memory.NewPostRepository()
	ctx := context.Background()

	authorID, _ := post_entity.NewAuthorID(testAuthorID)
	title, _ := post_entity.NewPostTitle("Hello")
	content, _ := post_entity.NewPostContent("World")
	post := post_entity.NewPost(authorID, title, content)

	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, post.ID())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID().String() != post.ID().String() {
		t.Error("id mismatch")
	}
}

func TestMemoryPostRepo_NotFound(t *testing.T) {
	repo := memory.NewPostRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, post_entity.NewPostID())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var nfe *customerror.NotFoundError
	if !asError(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestMemoryPostRepo_List(t *testing.T) {
	repo := memory.NewPostRepository()
	ctx := context.Background()

	authorID, _ := post_entity.NewAuthorID(testAuthorID)

	for i := range 5 {
		title, _ := post_entity.NewPostTitle(titleN(i))
		content, _ := post_entity.NewPostContent("Content")
		if err := repo.Create(ctx, post_entity.NewPost(authorID, title, content)); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	posts, _, err := repo.List(ctx, 3, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("expected 3, got %d", len(posts))
	}
}

func TestMemoryPostRepo_Update(t *testing.T) {
	repo := memory.NewPostRepository()
	ctx := context.Background()

	authorID, _ := post_entity.NewAuthorID(testAuthorID)
	title, _ := post_entity.NewPostTitle("Before")
	content, _ := post_entity.NewPostContent("Content")
	post := post_entity.NewPost(authorID, title, content)
	repo.Create(ctx, post) //nolint

	post.ToggleComments(false)
	if err := repo.Update(ctx, post); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := repo.GetByID(ctx, post.ID())
	if got.CommentsEnabled() {
		t.Error("expected comments disabled after update")
	}
}

func TestMemoryCommentRepo_CreateAndGet(t *testing.T) {
	repo := memory.NewCommentRepository()
	ctx := context.Background()

	c, err := comment_entity.NewComment(
		comment_entity.PostID(post_entity.NewPostID().String()),
		comment_entity.AuthorID(testAuthorID),
		nil,
		"Hello comment",
	)
	if err != nil {
		t.Fatalf("new comment: %v", err)
	}

	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got.Text()) != "Hello comment" {
		t.Errorf("text mismatch")
	}
}

func TestMemoryCommentRepo_ListByPostID(t *testing.T) {
	repo := memory.NewCommentRepository()
	ctx := context.Background()

	postID := comment_entity.PostID(post_entity.NewPostID().String())
	postEntityID, _ := post_entity.ParsePostID(string(postID))

	for i := range 4 {
		c, _ := comment_entity.NewComment(postID, comment_entity.AuthorID(testAuthorID), nil, comment_entity.CommentContent(titleN(i)))
		repo.Create(ctx, c)
	}

	comments, _, err := repo.ListByPostID(ctx, postEntityID, 10, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(comments) != 4 {
		t.Errorf("expected 4, got %d", len(comments))
	}
}

func TestMemoryCommentRepo_HierarchyAndReplies(t *testing.T) {
	repo := memory.NewCommentRepository()
	ctx := context.Background()

	postID := comment_entity.PostID(post_entity.NewPostID().String())
	postEntityID, _ := post_entity.ParsePostID(string(postID))

	root, _ := comment_entity.NewComment(postID, comment_entity.AuthorID(testAuthorID), nil, "Root")
	repo.Create(ctx, root)

	for i := range 3 {
		c, _ := comment_entity.NewComment(postID, comment_entity.AuthorID(testAuthorID), ptr(root.ID()), comment_entity.CommentContent(titleN(i)))
		repo.Create(ctx, c)
	}

	// Корневых должно быть 1
	roots, _, err := repo.ListByPostID(ctx, postEntityID, 10, "")
	if err != nil {
		t.Fatalf("list root: %v", err)
	}
	if len(roots) != 1 {
		t.Errorf("expected 1 root comment, got %d", len(roots))
	}

	// Ответов должно быть 3
	replies, _, err := repo.ListByParentID(ctx, root.ID(), 10, "")
	if err != nil {
		t.Fatalf("list replies: %v", err)
	}
	if len(replies) != 3 {
		t.Errorf("expected 3 replies, got %d", len(replies))
	}
}

func TestMemoryCommentRepo_BatchGetByPostIDs(t *testing.T) {
	repo := memory.NewCommentRepository()
	ctx := context.Background()

	pid1 := post_entity.NewPostID()
	pid2 := post_entity.NewPostID()

	for i := range 2 {
		c, _ := comment_entity.NewComment(comment_entity.PostID(pid1.String()), comment_entity.AuthorID(testAuthorID), nil, comment_entity.CommentContent(titleN(i)))
		repo.Create(ctx, c)
	}
	c, _ := comment_entity.NewComment(comment_entity.PostID(pid2.String()), comment_entity.AuthorID(testAuthorID), nil, "Single")
	repo.Create(ctx, c)

	result, err := repo.BatchGetByPostIDs(ctx, []post_entity.PostID{pid1, pid2})
	if err != nil {
		t.Fatalf("batch get: %v", err)
	}
	if len(result[pid1]) != 2 {
		t.Errorf("pid1: expected 2 comments, got %d", len(result[pid1]))
	}
	if len(result[pid2]) != 1 {
		t.Errorf("pid2: expected 1 comment, got %d", len(result[pid2]))
	}
}
