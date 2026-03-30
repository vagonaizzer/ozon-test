package post

import (
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	ErrEmptyTitle     = errors.New("title cannot be empty")
	ErrTitleTooLong   = errors.New("title is too long")
	ErrEmptyContent   = errors.New("content cannot be empty")
	ErrContentTooLong = errors.New("content is too long")
)

const (
	MaxTitleLength   = 255
	MaxContentLength = 10000
)

type Post struct {
	id              PostID
	authorID        AuthorID
	title           PostTitle
	content         PostContent
	commentsEnabled bool
	createdAt       time.Time
}

type PostID struct{ value ulid.ULID }
type AuthorID struct{ value ulid.ULID }
type PostTitle struct{ v string }
type PostContent struct{ v string }

func NewPost(authorID AuthorID, title PostTitle, content PostContent) *Post {
	return &Post{
		id:              NewPostID(),
		authorID:        authorID,
		title:           title,
		content:         content,
		commentsEnabled: true,
		createdAt:       time.Now().UTC(),
	}
}

func (id AuthorID) String() string { return id.value.String() }

func NewPostTitle(v string) (PostTitle, error) {
	if v == "" {
		return PostTitle{}, ErrEmptyTitle
	}
	if len([]rune(v)) > MaxTitleLength {
		return PostTitle{}, ErrTitleTooLong
	}
	return PostTitle{v: v}, nil
}

func NewPostContent(v string) (PostContent, error) {
	if v == "" {
		return PostContent{}, ErrEmptyContent
	}
	if len([]rune(v)) > MaxContentLength {
		return PostContent{}, ErrContentTooLong
	}
	return PostContent{v: v}, nil
}

func (p *Post) ID() PostID            { return p.id }
func (p *Post) AuthorID() AuthorID    { return p.authorID }
func (p *Post) Title() PostTitle      { return p.title }
func (p *Post) Content() PostContent  { return p.content }
func (p *Post) CommentsEnabled() bool { return p.commentsEnabled }
func (p *Post) CreatedAt() time.Time  { return p.createdAt }

func generatePostID() PostID         { return PostID{value: ulid.Make()} }
func NewPostID() PostID              { return generatePostID() }
func (id PostID) String() string     { return id.value.String() }
func (c PostContent) String() string { return c.v }
func (t PostTitle) String() string   { return t.v }

// ParsePostID восстанавливает PostID из строки (ULID), используется репозиторием.
func ParsePostID(s string) (PostID, error) {
	u, err := ulid.ParseStrict(s)
	if err != nil {
		return PostID{}, err
	}
	return PostID{value: u}, nil
}

// NewAuthorID создаёт AuthorID из строки (ULID).
func NewAuthorID(s string) (AuthorID, error) {
	u, err := ulid.ParseStrict(s)
	if err != nil {
		return AuthorID{}, err
	}
	return AuthorID{value: u}, nil
}

func (p *Post) ToggleComments(enabled bool) {
	p.commentsEnabled = enabled
}

// RestorePost перезаписывает мутабельные поля Post значениями из хранилища.
// Используется только репозиториями — не для бизнес-логики.
func RestorePost(p *Post, id PostID, authorID AuthorID, commentsEnabled bool, createdAt time.Time) {
	p.id = id
	p.authorID = authorID
	p.commentsEnabled = commentsEnabled
	p.createdAt = createdAt
}
