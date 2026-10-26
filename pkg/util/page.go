package util

// PageSlice 内存分页：pageNo/pageSize <= 0 返回全部，页码越界返回空切片。
func PageSlice[T any](list []T, pageNo, pageSize int) []T {
	if pageNo <= 0 || pageSize <= 0 {
		return list
	}

	total := int64(len(list))
	start := (int64(pageNo) - 1) * int64(pageSize)
	if start < 0 || start >= total {
		return []T{}
	}

	end := start + int64(pageSize)
	if end > total {
		end = total
	}
	return list[int(start):int(end)]
}
