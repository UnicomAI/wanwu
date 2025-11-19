package util

func StrToBool(str string) bool {
	return str == "true"
}

// HasDuplicate 判断两个数组是否有交集时间复杂度o(m*n) [EN] HasDuplicate determines whether two arrays have intersection time complexity o(m*n)
func HasDuplicate(a, b []string) bool {
	for _, itemA := range a {
		for _, itemB := range b {
			if itemA == itemB {
				return true
			}
		}
	}
	return false
}

// HasIntersection 判断两个数组是否有交集时间复杂度o(m+n) [EN] HasIntersection determines whether two arrays intersect. Time complexity o(m+n)
func HasIntersection(arr1, arr2 []string) bool {
	// 优化：将较小的数组放入 map 中 [EN] Optimization: Put smaller arrays into map
	if len(arr1) > len(arr2) {
		arr1, arr2 = arr2, arr1
	}

	// 使用空结构体节省内存 [EN] Use empty structures to save memory
	set := make(map[string]bool, len(arr1))
	for _, s := range arr1 {
		set[s] = true
	}

	// 检查交集 [EN] Check intersection
	for _, s := range arr2 {
		if set[s] {
			return true
		}
	}

	return false
}
