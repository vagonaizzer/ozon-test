package comment

import (
	"errors"
	"sync/atomic"
	"time"
)

var commentIDCounter int64

var (
	ErrEmptyField     = errors.New("comment content cannot be empty")
	ErrCommentTooLong = errors.New("comment text exceeds maximum length")
)

const MaxTextLength = 2000

type (
	CommentID      int64
	PostID         string
	AuthorID       string
	CommentContent string
)

type Comment struct {
	commentID      CommentID
	postID         PostID
	authorID       AuthorID
	parentID       *CommentID
	commentContent CommentContent
	createdAt      time.Time
}

func NewComment(
	postID PostID,
	authorID AuthorID,
	parentID *CommentID,
	content CommentContent,
) (*Comment, error) {
	c := &Comment{
		commentID:      NewCommentID(),
		postID:         postID,
		authorID:       authorID,
		parentID:       parentID,
		commentContent: content,
		createdAt:      time.Now().UTC(),
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func NewCommentID() CommentID {
	return CommentID(atomic.AddInt64(&commentIDCounter, 1))
}

func (c *Comment) Validate() error {
	if c.commentContent == "" {
		return ErrEmptyField
	}
	if len([]rune(string(c.commentContent))) > MaxTextLength {
		return ErrCommentTooLong
	}
	return nil
}

func (c *Comment) ID() CommentID        { return c.commentID }
func (c *Comment) PostID() PostID       { return c.postID }
func (c *Comment) AuthorID() AuthorID   { return c.authorID }
func (c *Comment) ParentID() *CommentID { return c.parentID }
func (c *Comment) Text() CommentContent { return c.commentContent }
func (c *Comment) CreatedAt() time.Time { return c.createdAt }

func (c *Comment) SetID(id CommentID) { c.commentID = id }

func (c *Comment) IsRoot() bool { return c.parentID == nil }

func (c *Comment) UpdateContent(newContent CommentContent) error {
	old := c.commentContent
	c.commentContent = newContent
	if err := c.Validate(); err != nil {
		c.commentContent = old
		return err
	}
	return nil
}
