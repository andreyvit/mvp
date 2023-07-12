package mvp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(fmt.Errorf("cannot read %d random bytes: %w", len, err))
	}
	return b
}

func RandomHex(len int) string {
	b := RandomBytes((len + 1) / 2)
	return hex.EncodeToString(b)[:len]
}

func RandomAlpha(n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 characters
	raw := RandomBytes(n)
	for i, b := range raw {
		raw[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(raw)
}

func RandomDigits(n int) string {
	const alphabet = "0123456789"
	raw := RandomBytes(n)
	for i, b := range raw {
		raw[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(raw)
}

func WriteFileAtomic(path string, data []byte, perm fs.FileMode) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".~"+filepath.Base(path)+".*")
	if err != nil {
		return err
	}

	var ok, closed bool
	defer func() {
		if !closed {
			temp.Close()
		}
		if !ok {
			os.Remove(temp.Name())
		}
	}()

	err = temp.Chmod(perm)
	if err != nil {
		return err
	}

	_, err = temp.Write(data)
	if err != nil {
		return err
	}

	err = temp.Close()
	closed = true
	if err != nil {
		return err
	}

	err = os.Rename(temp.Name(), path)
	if err != nil {
		return err
	}

	ok = true
	return nil
}

// TrimPort removes :port part (if any) from the given string and returns just the hostname.
func TrimPort(host string) string {
	h, _, _ := net.SplitHostPort(host)
	if h == "" {
		return host
	}
	return h
}

func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 UTC")
	w.Header().Set("Cache-Control", "no-cache, no-store, no-transform, must-revalidate, private, max-age=0")
	w.Header().Set("Pragma", "no-cache")
}

func MarkPublicImmutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

func MarkPublicMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, no-cache, max-age=0")
}

func MarkPrivateMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-cache, max-age=0")
}

func DetermineMIMEType(r *http.Request) string {
	s := r.Header.Get("Content-Type")
	if s == "" {
		return ""
	}
	ctype, _, err := mime.ParseMediaType(s)
	if ctype == "" || err != nil {
		return ""
	}
	return ctype
}

type action func()

func (_ action) String() string {
	return ""
}

func (_ action) IsBoolFlag() bool {
	return true
}

func (f action) Set(string) error {
	f()
	os.Exit(0)
	return nil
}

// CanonicalEmail returns an email suitable for unique checks.
func CanonicalEmail(email string) string {
	email = strings.TrimSpace(email)
	return strings.ToLower(email)
}

// CanonicalPhone returns an phone number suitable for unique checks.
func CanonicalPhone(number string) string {
	return strings.Map(cleanupPhoneRune, number)
}

func cleanupPhoneRune(r rune) rune {
	if r >= '0' && r <= '9' {
		return r
	} else if r == '+' {
		return r
	} else {
		return -1
	}
}

// EmailRateLimitingKey is a slightly paranoid function that maps emails into string keys to use for rate limiting.
func EmailRateLimitingKey(email string) string {
	username, host, found := strings.Cut(email, "@")
	if !found {
		return "invalid" // use single key for all invalid emails to cut down on stupid shenanigans
	}
	// get rid of local part (after +)
	username, _, _ = strings.Cut(username, "+")
	username = strings.Map(keepOnlyLettersAndNumbers, username)
	email = username + "@" + host
	email = strings.ToLower(email)
	return email
}

var lettersAndNumbers = [128]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}

func keepOnlyLettersAndNumbers(r rune) rune {
	if r < 128 && lettersAndNumbers[r] != 0 {
		return r
	} else {
		return -1
	}
}

func isWhitespaceOrComma(r rune) bool {
	return r == ' ' || r == ','
}

func JoinClasses(items ...any) string {
	var classes []string
	for _, item := range items {
		switch item := item.(type) {
		case []string:
			if classes == nil {
				classes = item
			} else {
				classes = AddClassList(classes, item)
			}
		case string:
			classes = AddClasses(classes, item)
		default:
			panic(fmt.Errorf("JoinClasses: invalid item %T %v", item, item))
		}
	}
	return JoinClassList(classes)
}

func JoinClassStrings(items ...string) string {
	var classes []string
	for _, item := range items {
		classes = AddClasses(classes, item)
	}
	return JoinClassList(classes)
}

func JoinClassList(classes []string) string {
	return strings.Join(classes, " ")
}

func AddClasses(classes []string, items string) []string {
	if len(items) == 0 {
		return classes
	}
	return AddClassList(classes, strings.Fields(items))
}

func AddClassList(classes []string, items []string) []string {
	if len(items) == 0 {
		return classes
	}
	for _, c := range items {
		classes = AddSingleClass(classes, c)
	}
	return classes
}

func AddSingleClass(classes []string, item string) []string {
	if len(item) == 0 {
		return classes
	}
	if s, ok := strings.CutPrefix(item, "remove:"); ok {
		if i := slices.Index(classes, s); i >= 0 {
			return slices.Delete(classes, i, i+1)
		}
	} else {
		if i := slices.Index(classes, s); i < 0 {
			return append(classes, item)
		}
	}
	return classes
}

func Stringify(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	case template.HTML:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func HTMLify(v any) template.HTML {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return HTMLifyString(v)
	case template.HTML:
		return v
	default:
		return HTMLifyString(fmt.Sprint(v))
	}
}

func HTMLifyString(text string) template.HTML {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = template.HTMLEscapeString(line)
	}
	return template.HTML(strings.Join(lines, "<br>\n"))
}

func sendSignal(c chan<- struct{}) {
	c <- struct{}{}
}

func ensureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	} else {
		return s + suffix
	}
}

func JoinInlineCSS(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" {
		return b
	} else if b == "" {
		return a
	} else if strings.HasSuffix(a, ";") {
		return a + " " + b
	} else {
		return a + "; " + b
	}
}

var (
	mobileUARe  = regexp.MustCompile(`(?i)(iPhone|Android)`)
	androidUARe = regexp.MustCompile(`(?i)Android`)
)

func IsMobileUA(ua string) bool {
	return mobileUARe.MatchString(ua)
}
func IsAndroidUA(ua string) bool {
	return androidUARe.MatchString(ua)
}
func IsMobileRequest(r *http.Request) bool {
	return IsMobileUA(r.Header.Get("User-Agent"))
}
func IsAndroidRequest(r *http.Request) bool {
	return IsAndroidUA(r.Header.Get("User-Agent"))
}
func SMSLinkURI(phone, body string, ua string) string {
	qs := PlusToPercent20(url.Values{"body": {body}}.Encode())
	if IsAndroidUA(ua) {
		return "sms://" + phone + "/?" + qs
	} else {
		return "sms://" + phone + "/&" + qs
	}
}
func TweetIntentURL(body string, linkURL string) string {
	q := url.Values{
		"text": {body},
	}
	if linkURL != "" {
		q["url"] = []string{linkURL}
	}
	return "https://twitter.com/intent/tweet?" + q.Encode()
}
func WhatsappSendURL(body string, ua string) string {
	if IsMobileUA(ua) {
		q := url.Values{
			"text": {body},
		}
		return "whatsapp://send/?" + q.Encode()
	} else {
		q := url.Values{
			"text": {body},
		}
		return "https://api.whatsapp.com/send/?" + q.Encode()
	}
}

func PlusToPercent20(s string) string {
	return strings.ReplaceAll(s, "+", "%20")
}
