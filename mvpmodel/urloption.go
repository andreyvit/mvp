package mvpm

import "strings"

type URLOption uint64

var (
	urlOptionCount int
	urlOptionNames = make(map[URLOption]string, 64)
)

func NewURLOption(name string) URLOption {
	if urlOptionCount == 64 {
		panic("out of URL options")
	}
	i := urlOptionCount
	urlOptionCount++
	v := URLOption(1 << i)
	urlOptionNames[v] = name
	return v
}

func (v URLOption) Contains(c URLOption) bool {
	return (v & c) == c
}

func (v URLOption) String() string {
	if v == 0 {
		return "default"
	}
	var buf strings.Builder
	for i := 0; i < 64; i++ {
		c := URLOption(1 << i)
		if v.Contains(c) {
			if v == c {
				return urlOptionNames[c]
			}
			if buf.Len() > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(urlOptionNames[c])
		}
	}
	return buf.String()
}
