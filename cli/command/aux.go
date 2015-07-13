package command

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ksarch-saas/cc/frontend/api"
)

var Put = fmt.Println
var Putf = fmt.Printf

func ShowResponse(resp *api.Response) {
	if resp.Errno == 0 {
		if resp.Body == nil {
			fmt.Println(resp.Errmsg)
		} else {
			var out bytes.Buffer
			data, _ := json.Marshal(resp.Body)
			json.Indent(&out, []byte(data), "", "  ")
			fmt.Printf("%s: %#v\n", resp.Errmsg, out.String())
		}
	} else {
		fmt.Println("Command failed:", resp.Errmsg)
	}
}
