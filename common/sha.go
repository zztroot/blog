package common

import (
	"crypto/sha1"
	"fmt"
)

func ToSha1String(str string)string{
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func ToSha1Byte(str string)[]byte{
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return bs
}
