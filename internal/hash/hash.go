package hash

import (
	"hash/crc32"
)

const NumHashSlots = 16384

func GetHashSlot(key string) uint32 {
	crc32Value := crc32.ChecksumIEEE([]byte(key))

	hashSlot := crc32Value % NumHashSlots

	return hashSlot
}

func CalculateHashSlotRanges(numNodes int, numSlots int) map[int][]int {
	ranges := make(map[int][]int)
	slotsPerNode := numSlots / numNodes
	remainder := numSlots % numNodes

	start := 0
	for i := 1; i <= numNodes; i++ {
		end := start + slotsPerNode - 1
		if remainder > 0 {
			end++
			remainder--
		}
		ranges[i] = []int{start, end}
		start = end + 1
	}

	return ranges
}
