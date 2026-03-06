package util

import "fmt"

func ParsePagination(offsetStr, limitStr string) (int, int, error) {
	offset := 0
	limit := 50
	if offsetStr != "" {
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("%w: invalid offset", ErrValidation)
		}
	}
	if limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit <= 0 {
			return 0, 0, fmt.Errorf("%w: invalid limit", ErrValidation)
		}
	}
	return offset, limit, nil
}

func PaginateSlice[T any](items []T, offset, limit int) []T {
	if offset >= len(items) {
		return []T{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
