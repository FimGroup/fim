package tools

import (
	"crypto/rand"
	"encoding/hex"
	"runtime"
)

func RandomString() string {
	data := make([]byte, 16)
	var e error
	for i := 0; i < 3; i++ {
		if _, err := rand.Read(data); err != nil {
			e = err
			runtime.Gosched()
			continue
		} else {
			return hex.EncodeToString(data)
		}
	}
	panic(e)
}
