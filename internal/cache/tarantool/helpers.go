package tarantool

func toInt64(value any) (int64, error) {
	switch n := value.(type) {
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		if uint64(n) > uint64(^uint64(0)>>1) {
			return 0, ErrInvalidCacheEntry
		}
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		if n > uint64(^uint64(0)>>1) {
			return 0, ErrInvalidCacheEntry
		}
		return int64(n), nil
	case float64:
		if n < 0 || n > float64(^uint64(0)>>1) {
			return 0, ErrInvalidCacheEntry
		}
		return int64(n), nil
	default:
		return 0, ErrInvalidCacheEntry
	}
}
