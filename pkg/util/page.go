package util

// PageSlice 内存分页：取 list 第 pageNo 页（从 1 开始），页码越界返回空切片
func PageSlice[T any](list []T, pageNo, pageSize int) []T {
	start := (pageNo - 1) * pageSize
	if start >= len(list) {
		return []T{}
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}
