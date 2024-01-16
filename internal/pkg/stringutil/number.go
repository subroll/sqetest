package stringutil

import (
	"bufio"
	"crypto/rand"
	"io"
	"sync"
)

var randomReaderPool = sync.Pool{New: func() interface{} {
	return bufio.NewReader(rand.Reader)
}}

const (
	randomNumberCharset    = "0123456789"
	randomNumberCharsetLen = 10
	randomNumberMaxByte    = 255 - (256 % randomNumberCharsetLen)
)

func RandomNumbers(length uint8) (string, error) {
	reader := randomReaderPool.Get().(*bufio.Reader)
	defer randomReaderPool.Put(reader)

	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	var i uint8 = 0

	for {
		_, err := io.ReadFull(reader, r)
		if err != nil {
			return "", err
		}

		for _, rb := range r {
			if rb > randomNumberMaxByte {
				continue
			}

			b[i] = randomNumberCharset[rb%randomNumberCharsetLen]
			i++

			if i == length {
				return string(b), nil
			}
		}
	}
}
