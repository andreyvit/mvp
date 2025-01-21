package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

var testKey = []byte("helloworld")

var testKeyPrivateRSA = must(x509.ParsePKCS8PrivateKey(head(pem.Decode([]byte(
	strings.Join([]string{
		"-----BEGIN PRIVATE KEY-----",
		"MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDfHW5/O0hegh0D",
		"L1K3p0PpAkB56m5ekn3yeoC6nO3BkuAlexMhq9O/OpCDWJ3O/pVRONa/rbRJ8Qc+",
		"dL0w2IaKPhmNRGo1zOyBChI5xhT1vBjnsyR3RNvVYxh5XPn6ly4dscAzg97aCHBr",
		"uoEpqKLP5OC4j6GMGF7/CWCHmjl8WfdDvay3rbc7GPSJoRE9+taTlxNdW3nG85g2",
		"BaDDx/pIAtQu2BcqRJ3UGcLqPCKpqVKKWWtmCRWZpfpVNEr4UimokaC/BM7lvAKu",
		"/tFQe30o86aC1CgVTrg5MxYgHxHIvO/J0nXh63YIJLvnpu/+Seb2D2R3Dypmcu1G",
		"XfDZnA0lAgMBAAECggEAAZpYXGFPJvVVWFwDLUmY2Hgz1rcN8u+nfadOpx3l1rcD",
		"lZFZn6t4hSOvvRnVDGh80c1lkZFMaHRAV+exIkQz9z30o19Yn1P+ZlDttOziN/+8",
		"HWdb6DVzmjKw5CenA3C0RpyrzlLwthf1ws20ttlneLSNaWtdaZW/56IoHOFJRMZB",
		"7CnNKNkyCOjYCgn+pGJ50FN+E8RVYd9PVSwdn0ytKI5cUO8RJg3bTyvVtBmlD3Oq",
		"IhKWgwFfZCb2OicIOe9uJo2eLg7G2JMohVZS7mvn186Gz8DL3RcepuYMF51aW9aW",
		"zwAkETYd9qSok+CUxBJ6u6vy3OgbjHdGiKOxL7MP8QKBgQDyGH9oQjHSP5A6EmYU",
		"slWhVOAMbQDAKGsJNNaWexVPzNaPJS7XnpJJwoG5RL8psXxFgEz3HxHLyxduLU6j",
		"rT0fJOe0QGFXuf7uI4HAacgTXqtQAl5WhPSLgyyEQBiAF77FfxQ6zWuoDiYHefUR",
		"rNOTXgR1SZ5al7OevifxqZBJzQKBgQDr7dzbcKtaGLBbYQKdAe4jro3ZCPX3OB6a",
		"heDyf0UKDIQmP415S4j9VKojmVZIBxi0vM+xfDjz+gsfza4rbyd7Epsd4J8rCmNn",
		"052LA2zbrIonaKjGHeitj9twns83Dv5bcxih1Q5YNgJkB5IqFVa3MXQIuGBX4MkY",
		"BoCUrZmYuQKBgQCKI2ZXhCXPdQuDx0nOF2/67WYmUPAztRxWFXs4RCUF8rie1zWi",
		"PM32HnFM2KhHTwm80peYDndmFI1bBakwhcIxiipX1MB2gR+wnDwGIZXTT5pqvd88",
		"eQLctE1rbPNN676kDH3ri5kZPHGApJssqbPUC7p5fjdIM/V+57v9DabSJQKBgDpy",
		"10BWDV1ouGgrBGa5T7HvUJzwJ19zu8E0YaIx/Xyb1TFUlUvzdqCsFOp01ndJqsk7",
		"7Yhe6g0naRIb9oY0J2fKGDuypjwXesECIAmMc6+Ic0GIICfUyQwWk5q2/Ub6o/Er",
		"9nJBQiiAWN9HMOLUHoOL8N8oLlYXDjqxgbFTwLWhAoGBAKfYBiv2Lew+TJgRCtTU",
		"vAFxCyj6hT9/cPVINCoeHx/IIMBVJPXmmWkA1pLao6rbHVIZ1rEzOVJDmAnKIaA/",
		"WhMm9ABee+w4WnWGqmCZnum12K0k1YLSKPAtFxl/0tAp6TyEDR3Mj47Dz5TgWj2J",
		"w5Giyno3j4ksho63Yo8wFg1D",
		"-----END PRIVATE KEY-----"}, "\n")))).Bytes)).(*rsa.PrivateKey)

var testKeyPublicRSA = must(x509.ParsePKIXPublicKey(head(pem.Decode([]byte(
	strings.Join([]string{
		"-----BEGIN PUBLIC KEY-----",
		"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3x1ufztIXoIdAy9St6dD",
		"6QJAeepuXpJ98nqAupztwZLgJXsTIavTvzqQg1idzv6VUTjWv620SfEHPnS9MNiG",
		"ij4ZjURqNczsgQoSOcYU9bwY57Mkd0Tb1WMYeVz5+pcuHbHAM4Pe2ghwa7qBKaii",
		"z+TguI+hjBhe/wlgh5o5fFn3Q72st623Oxj0iaERPfrWk5cTXVt5xvOYNgWgw8f6",
		"SALULtgXKkSd1BnC6jwiqalSillrZgkVmaX6VTRK+FIpqJGgvwTO5bwCrv7RUHt9",
		"KPOmgtQoFU64OTMWIB8RyLzvydJ14et2CCS756bv/knm9g9kdw8qZnLtRl3w2ZwN",
		"JQIDAQAB",
		"-----END PUBLIC KEY-----"}, "\n")))).Bytes)).(*rsa.PublicKey)

var testNow = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		payload  string
		newToken string
		verifier Verifier
		signer   Signer
	}{
		{
			"HS256: jwt.io",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.dD_HjcF4ZoXwMj6Ov7q7uDqCZLeNMhOwC52WEGEG7P0",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.pBtuSBkUUz0-RMxWpH-uWr-4_C-AJiImWHiE7zxbcI4",
			NewHS256Verifier(testKey, ""),
			NewHS256Signer(testKey, ""),
		},
		{
			"HS256: jwt.io re-encoded via this library",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.pBtuSBkUUz0-RMxWpH-uWr-4_C-AJiImWHiE7zxbcI4",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"",
			NewHS256Verifier(testKey, ""),
			NewHS256Signer(testKey, ""),
		},
		{
			"RS256: jwt.io",
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.3K5k4K_8vb1WiFJZu0qlvjRKObvg7pLE9FjWUZR-vjXetfhkmteR6EjLH9d3fWJlrHzpi6FwhgMR573pUYyxQmxsJeWNWhHeltPy5HQ6DDUxVguH5EAryH4LuOP-87kjrNAo7xY0BO_cLFfoeL9SI8qQH8P3-TBESAhi26L62MipiAjh1ABqkCevZkmEFxAukJoKvEVipVoGGhV9RgyA_6YR1rlJVD-n6g4UwexmdAVJSaF1ggbeA6qVuLp1OAzArAJ93LhycPZ9DrlhDj6E2vOqXsPDH25jIPwYzRiJzIS3w0hLoKPbYyReN-aYZRwSIG8YEmkubRvaiAyx6aO6dg",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.0jlQy1NM1zpJxsb5ZDz53Qwl-Eu5EkJQ4krrwDKCFiT5VjS00dEiIfHZW8uhqsmvUQm5N1IwzKaiy6dbdVLWzFwIgAV-i3z6xLkCPJQ3R4LVlEQDPKEK-O90CqAlvq10uE8XQNpcKPjcT5m7TFpAhQXL4KrTXyYCrbKki1sAhTvMmwmBo-rrDR-yCGuvJJEgPp0qJxbRtXdR3Yea8CL4F8GT3QkaroP0oAIpJkQNlg3WoSvlqZpnHh4gKGJuYBnCWONM9TOVl-jk4Ebabze-LGtMoCTqihT43GI3-T9YcBP5Jqgm4JvZZDC7Emqcya645FDmZpt7FxBeIa4Kv7PvrA",
			NewRS256Verifier(testKeyPublicRSA, ""),
			NewRS256Signer(testKeyPrivateRSA, ""),
		},
		{
			"RS256: jwt.io re-encoded via this library",
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MTYyMzkwMjIsIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.0jlQy1NM1zpJxsb5ZDz53Qwl-Eu5EkJQ4krrwDKCFiT5VjS00dEiIfHZW8uhqsmvUQm5N1IwzKaiy6dbdVLWzFwIgAV-i3z6xLkCPJQ3R4LVlEQDPKEK-O90CqAlvq10uE8XQNpcKPjcT5m7TFpAhQXL4KrTXyYCrbKki1sAhTvMmwmBo-rrDR-yCGuvJJEgPp0qJxbRtXdR3Yea8CL4F8GT3QkaroP0oAIpJkQNlg3WoSvlqZpnHh4gKGJuYBnCWONM9TOVl-jk4Ebabze-LGtMoCTqihT43GI3-T9YcBP5Jqgm4JvZZDC7Emqcya645FDmZpt7FxBeIa4Kv7PvrA",
			`{"iat":1516239022,"name":"John Doe","sub":"1234567890"}`,
			"",
			NewRS256Verifier(testKeyPublicRSA, ""),
			NewRS256Signer(testKeyPrivateRSA, ""),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var token Token
			err := token.ParseString(test.input)
			if err != nil {
				t.Fatal(err)
			}
			err = token.ValidateWith(test.verifier)
			if err != nil {
				t.Fatal(err)
			}
			c := token.Claims()
			p := string(must(json.Marshal(c)))
			if p != test.payload {
				t.Errorf("Decode = %s, wanted %s", p, test.payload)
			}

			err = c.ValidateTimeAt(0, testNow)
			if err != nil {
				t.Fatal(err)
			}

			output, _ := SignString(c, nil, test.signer)
			if test.newToken == "" {
				test.newToken = test.input
			}
			if output != test.newToken {
				t.Errorf("** Sign = %q, wanted %q", output, test.newToken)
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

func head[T any](v T, _ any) T {
	return v
}
