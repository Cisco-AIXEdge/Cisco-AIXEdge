package internals

import (
	"fmt"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/providers"
)

func (c *Client) Interactive() {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("API key non-existent. Please do copilot-cfg <LLM Provider> <Model> <API KEY>")
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
	a.Interactive(cfg.SerialNumber)
	// c.Interactive_Telemetry()
}
