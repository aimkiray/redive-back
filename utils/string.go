package utils

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

func HashName(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func InList(l []string, n string) bool {
	for _, v := range l {
		if v == n {
			return true
		}
	}
	return false
}

func GetRandom() string {
	return fmt.Sprintf("%10v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000000000))
}
