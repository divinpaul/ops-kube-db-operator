package core

import (
	"crypto/rand"
	"fmt"
	"strings"
	"unicode"
)

// genPasswords generates password of length l
func GenPasswords(l int) (*string, error) {

	p, err := genPassword(l)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// genPassword generates a password of length l
func genPassword(l int) (string, error) {
	var p string
	var err error

	// url non-encode characters: [0-9a-zA-Z$.+!*(),_-]
	allow := &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x0030, 0x0039, 1}, // '0' - '9'
			{0x0021, 0x0021, 1}, // '!' - '!'
			{0x0024, 0x0024, 1}, // '$' - '$'
			{0x0028, 0x002e, 1}, // '(' - '.'
			{0x0041, 0x005a, 1}, // 'A' - 'Z'
			{0x005f, 0x005f, 1}, // '_' - '_'
			{0x0061, 0x007a, 1}, // 'a' - 'z'
		},
	}

	strip := func(r rune) rune {
		if unicode.Is(allow, r) {
			return r
		}
		return -1
	}

	for len(p) < l {
		b := make([]byte, l*2)
		_, err = rand.Read(b)
		if err != nil {
			err = fmt.Errorf("generate password: %v", err)
			return "", err
		}

		p += strings.Map(strip, string(b))
	}
	p = p[:l]
	return p, nil
}
