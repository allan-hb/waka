package player

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

var (
	random = rand.New(rand.NewSource(time.Now().Unix()))
)

func buildToken(uid string) string {
	s := fmt.Sprintf("%s%d%d", uid, time.Now().Unix(), random.Int63())
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}
