package data

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidIntegerValue = errors.New("invalid integer value")

func StringToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func ParseIntSlice[T ~int | ~int32 | ~int64](values []string) ([]T, error) {
	var result []T
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, ErrInvalidIntegerValue
		}
		result = append(result, T(i))
	}
	return result, nil
}

