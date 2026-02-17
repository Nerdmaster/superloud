package util

import (
	"crypto/md5"
	"fmt"
	"math/big"
)

// UserHash computes a semi-unique hash from a Discord user ID.
// Returns int64 derived from the MD5 of the user ID string.
func UserHash(userID string) int64 {
	hash := md5.Sum([]byte(userID))
	hexStr := fmt.Sprintf("%x", hash)

	n := new(big.Int)
	n.SetString(hexStr, 16)

	return n.Int64()
}
