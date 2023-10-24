package mvputil

import (
	"fmt"
	"net/url"
)

func CopyURLValues(dest, src url.Values) {
	for k, vv := range src {
		if len(vv) == 0 {
			delete(dest, k)
		} else {
			dest[k] = vv
		}
	}
}

func URLWithValues(urlStr string, values url.Values) string {
	if len(values) == 0 {
		return urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		panic(fmt.Errorf("invalid redirect URL: %q", urlStr))
	}

	q := u.Query()
	CopyURLValues(q, values)
	u.RawQuery = q.Encode()

	return u.String()
}
