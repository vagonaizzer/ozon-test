package tests

import (
	"errors"
	"fmt"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
)

const testAuthorID = "01JQHX00000000000000000001"

func ptr[T any](v T) *T { return &v }

// titleN генерирует заголовок вида "Title 0", "Title 1", …
func titleN(n int) string { return fmt.Sprintf("Title %d", n) }

// asError проверяет, можно ли распаковать err в *T через errors.As.
func asError[T error](err error, target *T) bool {
	return errors.As(err, target)
}

// Убеждаемся, что comment_entity используется в пакете тестов.
var _ = comment_entity.MaxTextLength
