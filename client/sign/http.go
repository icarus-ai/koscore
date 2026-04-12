package sign

import (
	"encoding/json"

	"github.com/kernel-ai/koscore/utils/http"
	"github.com/kernel-ai/koscore/utils/types"
)

/*
func httpGet[T any](uri string, heads types.MapSS) (target T, e error) {
	var data []byte
	heads["Content-Type"] = "application/json"
	if data, e = http.Get(uri, heads); e != nil { return }
	if       e = json.Unmarshal(data, &target); e != nil { return }
	return
}
*/

func httpPost[T any](uri string, data []byte, heads types.MapSS) (target T, e error) {
	heads["Content-Type"] = "application/json"
	if data, e = http.Post(uri, data, heads); e != nil {
		return
	}
	if e = json.Unmarshal(data, &target); e != nil {
		return
	}
	return
}
