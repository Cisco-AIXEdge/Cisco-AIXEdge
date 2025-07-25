package internals

import (
	"fmt"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/providers"
)

func (c *Client) Prompt(content string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("API key non-existent. Please do aixedge-cfg <LLM Provider> <Model> <API KEY>")
		}
	}()
	cfg, err := c.configRead()
	if err != nil {
		panic(err)
	}
	engine := providers.Engine{
		Provider: cfg.Engine,
		Version:  cfg.EngineVERSION,
	}
	a := providers.Client{
		API:    cfg.Apikey,
		Engine: engine,
	}
	a.Prompt(content)

}
