package jwt

import (
	"encoding/json"
	"testing"
	"time"
)

var testKey = []byte("helloworld")

var testNow = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		payload  string
		newToken string
	}{
		{
			"jwt.io",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.dD_HjcF4ZoXwMj6Ov7q7uDqCZLeNMhOwC52WEGEG7P0",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.pBtuSBkUUz0-RMxWpH-uWr-4_C-AJiImWHiE7zxbcI4",
		},
		{
			"jwt.io re-encoded via this library",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.pBtuSBkUUz0-RMxWpH-uWr-4_C-AJiImWHiE7zxbcI4",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var token Token
			err := token.ParseString(test.input)
			if err != nil {
				t.Fatal(err)
			}
			err = token.ValidateHS256(testKey)
			if err != nil {
				t.Fatal(err)
			}
			c := token.Claims()
			p := string(must(json.Marshal(c)))
			if p != test.payload {
				t.Errorf("DecodeHS256 = %s, wanted %s", p, test.payload)
			}

			err = c.ValidateTimeAt(0, testNow)
			if err != nil {
				t.Fatal(err)
			}

			output := SignHS256String(c, nil, testKey)
			if test.newToken == "" {
				test.newToken = test.input
			}
			if output != test.newToken {
				t.Errorf("** SignHS256 = %q, wanted %q", output, test.newToken)
			}
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
