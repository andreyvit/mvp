package forms

import "strings"

func JoinNames(path ...string) string {
	var buf strings.Builder
	for _, s := range path {
		if s == "" {
			continue
		}
		if buf.Len() == 0 {
			buf.WriteString(s)
		} else {
			j := strings.IndexByte(s, '[')
			if j < 0 {
				buf.WriteByte('[')
				buf.WriteString(s)
				buf.WriteByte(']')
			} else {
				if j > 0 {
					buf.WriteByte('[')
					buf.WriteString(s[:j])
					buf.WriteByte(']')
				}
				buf.WriteString(s[j:])
			}
		}
	}
	return buf.String()
}

func SplitName(name string) []string {
	var path []string
	for {
		component, suffix, _ := strings.Cut(name, "[")
		if component != "" {
			path = append(path, component)
		}
		if suffix == "" {
			break
		}
		component, name, _ = strings.Cut(suffix, "]")
		if component != "" {
			path = append(path, component)
		}
	}
	return path
}
