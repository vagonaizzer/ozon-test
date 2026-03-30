package common

type Page[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
}
