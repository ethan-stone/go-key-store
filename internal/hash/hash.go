package hash

import "hash/crc32"

const numHashSlots = 16384

func GetHashSlot(key string) uint32 {
	crc32Value := crc32.ChecksumIEEE([]byte(key))

	hashSlot := crc32Value % numHashSlots

	return hashSlot
}
