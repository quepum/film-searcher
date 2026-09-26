package main

import "fmt"

func twoSum(nums []int, target int) []int {
	m := make(map[int]int)

	for i := range nums {
		needed := target - nums[i]
		if _, ok := m[needed]; ok {
			return []int{m[needed], i}
		}
		m[nums[i]] = i

	}
	return []int{}
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9
	fmt.Println(twoSum(nums, target))
}
