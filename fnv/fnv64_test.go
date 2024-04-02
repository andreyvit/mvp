package fnv

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	stdfnv "hash/fnv"
	"testing"
)

func TestFNV64(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{"", std64(nil)},
		{"hello", "a430d84680aabd0b"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test), func(t *testing.T) {
			sum := String64(test.Input)
			actual := sum.String()
			if actual != test.Expected {
				t.Errorf("** String64 -> got %s, wanted %s", actual, test.Expected)
			}

			raw := must(json.Marshal(sum))
			var sum2 Hash64
			ensure(json.Unmarshal(raw, &sum2))
			if sum2 != sum {
				t.Errorf("** json.Unmarshal -> got %v, wanted %v", sum2, sum)
			}
		})
	}
}

func std64(b []byte) string {
	sum := stdfnv.New64a()
	sum.Write(b)
	return hex.EncodeToString(sum.Sum(nil))
}

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
