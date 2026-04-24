package sign

import (
	"encoding/json"

	"github.com/kernel-ai/koscore/utils/http"
	"github.com/kernel-ai/koscore/utils/types"
	//"github.com/kernel-ai/koscore/utils/comm"
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

func http_post[T any](uri string, data []byte, heads types.MapSS) (target T, e error) {
	heads["Content-Type"] = "application/json"
	if data, e = http.Post(uri, data, heads); e != nil {
		return
	}
	if e = json.Unmarshal(data, &target); e != nil {
		return
	}
	return
}

/*
func http_post_debug[T any](uri string, data []byte, heads types.MapSS) (target T, e error) {
	heads["Content-Type"] = "application/json"
	comm.LOGW("sign: %s", uri)
	comm.LOGW("  > body: %s", string(data))
	if data, e = http.Post(uri, data, heads); e != nil {
		comm.LOGW("   > post: %v", e)
		return
	}
	comm.LOGW("   > raw: %s", string(data))
	//comm.Fwrite(fmt.Sprintf("%s/kosbot/src/bot/_bin/sign_%s_%d_bin", comm.RootSD, cmd, seq), data)
	if e = json.Unmarshal(data, &target); e != nil {
		comm.LOGW("   > json unmarshal: %v", e)
		return
	}
	if rsp, ok := any(target).(rsp_sign_v2); ok { comm.LOGW("  > Y: %X - %X - %s", rsp.Value.Sign, rsp.Value.Token, string(rsp.Value.Extra)) }
	return
}
*/
