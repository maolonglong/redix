// copied from https://github.com/gin-gonic/gin/blob/master/internal/bytesconv/bytesconv.go

package bytesconv

import "unsafe"

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			cap int
		}{s, len(s)},
	))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
