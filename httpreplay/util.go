package httpreplay

import (
	"encoding/json"
)

func mustMarshalString(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(raw)
}
