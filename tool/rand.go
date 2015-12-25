package tool

import (
	"math/rand"
	"strconv"
	"time"
)

//产生statr和end之间的随机数
func Rand(start, end int) (code string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(end)
	for ; num <= start; num = r.Intn(end) {
	}
	return strconv.Itoa(num)
}
