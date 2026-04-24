package sign

import (
	"strings"

	"github.com/kernel-ai/koscore/utils"
)

func (c *Client) GetStat() (ret string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, i := range c.instances {
		ret += utils.Ternary(i.latency.Load() < server_latency_down, "\non: ", "\noff: ")
		ret += c.GetSignHost(i.server)
	}
	if len(ret) > 0 {
		ret = ret[1:]
	}
	return
}

func (c *Client) GetSignHost(uri string) string {
	if idx := strings.Index(uri, "//"); idx > 0 {
		uri = uri[idx+2:]
	}
	if idx := strings.Index(uri, "/"); idx > 0 {
		uri = uri[:idx]
	}
	return uri
}
