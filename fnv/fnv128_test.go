package fnv

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	stdfnv "hash/fnv"
	"testing"
)

func TestFNV128(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{"", std128(nil)},
		{"hello", std128([]byte("hello"))},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test), func(t *testing.T) {
			sum := String128(test.Input)
			actual := sum.String()
			if actual != test.Expected {
				t.Errorf("** String128 -> got %s, wanted %s", actual, test.Expected)
			}

			raw := must(json.Marshal(sum))
			var sum2 Hash128
			ensure(json.Unmarshal(raw, &sum2))
			if sum2 != sum {
				t.Errorf("** json.Unmarshal -> got %v, wanted %v", sum2, sum)
			}
		})
	}
}

func std128(b []byte) string {
	sum := stdfnv.New128a()
	sum.Write(b)
	return hex.EncodeToString(sum.Sum(nil))
}
