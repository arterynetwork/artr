package util

import "strings"

func RemoveStringPreserveOrder(arr *[]string, item string) bool {
	for i, s := range *arr {
		if s == item {
			copy((*arr)[i:], (*arr)[i+1:])
			*arr = (*arr)[:len(*arr)-1]
			return true
		}
	}
	return false
}

func RemoveStringFast(arr *[]string, item string) bool {
	for i, s := range *arr {
		if s == item {
			(*arr)[i] = (*arr)[len(*arr)-1]
			*arr = (*arr)[:len(*arr)-1]
			return true
		}
	}
	return false
}

func AddStringSorted(arr *[]string, item string) bool {
	for i, s := range *arr {
		if s == item {
			return false
		}
		if s > item {
			*arr = append(*arr, "")
			copy((*arr)[i+1:], (*arr)[i:])
			(*arr)[i] = item
			return true
		}
	}
	*arr = append(*arr, item)
	return true
}

func MergeStringsSorted(arr1 []string, arr2 []string) []string {
	if len(arr2) == 0 {
		return arr1
	}
	if len(arr1) == 0 {
		return arr2
	}
	res := make([]string, len(arr1)+len(arr2))
	i1 := 0
	i2 := 0
	ir := 0
	for {
		if i1 == len(arr1) {
			ir += copy(res[ir:], arr2[i2:])
			return res[:ir]
		}
		if i2 == len(arr2) {
			ir += copy(res[ir:], arr1[i1:])
			return res[:ir]
		}
		switch cmp := strings.Compare(arr1[i1], arr2[i2]); true {
		case cmp < 0:
			res[ir] = arr1[i1]
			ir += 1
			i1 += 1
		case cmp == 0:
			res[ir] = arr1[i1]
			ir += 1
			i1 += 1
			i2 += 1
		case cmp > 0:
			res[ir] = arr2[i2]
			ir += 1
			i2 += 1
		}
	}
}

func ContainsString(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}
