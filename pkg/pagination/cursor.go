package pagination

import (
	"encoding/base64"
	"strconv"
)

const DefaultLimit = 20
const MaxLimit = 100


func EncodeCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}


func DecodeCursor(cursor string) string {
	if cursor == "" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return ""
	}
	return string(b)
}


func EncodeIntCursor(id int64) string {
	return EncodeCursor(strconv.FormatInt(id, 10))
}


func DecodeIntCursor(cursor string) int64 {
	s := DecodeCursor(cursor)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}


func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}
