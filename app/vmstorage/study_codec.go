package main

import (
	"bytes"
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding"
)

const (
	escapeChar       = 0
	tagSeparatorChar = 1
	kvSeparatorChar  = 2
)

func main() {
	var b []byte
	var c1 []byte

	b = encoding.MarshalUint32(b, 123)
	b = encoding.MarshalUint64(b, uint64(77777777))
	b = encoding.MarshalUint32(b, 456)

	fmt.Println(b)

	a := encoding.UnmarshalUint32(b)
	//将 b 中从下标 4 到 endIndex-1 下的元素创建为一个新的切片。
	b = b[4:]
	d := encoding.UnmarshalUint64(b)
	b = b[8:]
	c := encoding.UnmarshalUint32(b)

	fmt.Println(a, c, d)

	b = b[:0]
	b = marshalTagValue(b, []byte("111111abc"))
	fmt.Println(b)
	_, c1, _ = unmarshalTagValue(c1, b)

	fmt.Println(string(c1[:]))

}

func marshalTagValue(dst, src []byte) []byte {
	fmt.Println(src)
	n1 := bytes.IndexByte(src, escapeChar)
	n2 := bytes.IndexByte(src, tagSeparatorChar)
	n3 := bytes.IndexByte(src, kvSeparatorChar)
	if n1 < 0 && n2 < 0 && n3 < 0 {
		// Fast path.
		dst = append(dst, src...)
		dst = append(dst, tagSeparatorChar)
		return dst
	}

	// Slow path.
	for _, ch := range src {
		switch ch {
		case escapeChar:
			dst = append(dst, escapeChar, '0')
		case tagSeparatorChar:
			dst = append(dst, escapeChar, '1')
		case kvSeparatorChar:
			dst = append(dst, escapeChar, '2')
		default:
			dst = append(dst, ch)
		}
	}

	dst = append(dst, tagSeparatorChar)
	return dst
}

func unmarshalTagValue(dst, src []byte) ([]byte, []byte, error) {
	n := bytes.IndexByte(src, tagSeparatorChar)
	if n < 0 {
		return src, dst, fmt.Errorf("cannot find the end of tag value")
	}
	b := src[:n]
	src = src[n+1:]
	for {
		n := bytes.IndexByte(b, escapeChar)
		if n < 0 {
			dst = append(dst, b...)
			return src, dst, nil
		}
		dst = append(dst, b[:n]...)
		b = b[n+1:]
		if len(b) == 0 {
			return src, dst, fmt.Errorf("missing escaped char")
		}
		switch b[0] {
		case '0':
			dst = append(dst, escapeChar)
		case '1':
			dst = append(dst, tagSeparatorChar)
		case '2':
			dst = append(dst, kvSeparatorChar)
		default:
			return src, dst, fmt.Errorf("unsupported escaped char: %c", b[0])
		}
		b = b[1:]
	}
}
