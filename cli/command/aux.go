package command

import (
	"fmt"

	"github.com/jxwr/cc/frontend/api"
)

func ShowResponse(resp *api.Response) {
	if resp.Errno == 0 {
		if resp.Body == nil {
			fmt.Println(resp.Errmsg)
		} else {
			fmt.Println(resp.Errmsg, resp.Body)
		}
	} else {
		fmt.Println("Command failed:", resp.Errmsg)
	}
}
