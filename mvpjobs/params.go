package mvpjobs

import "encoding/json"

type Params interface {
	JobName() string
	SetJobName(name string)
}

func EncodeParams(in Params) []byte {
	if in != nil {
		return must(json.Marshal(in))
	} else {
		return []byte("{}")
	}
}

type NoParams struct{}

func (_ NoParams) JobName() string        { return "" }
func (_ NoParams) SetJobName(name string) {}
