package internals

import (
	"fmt"

	"github.com/Cisco-AIXEdge/Cisco-AIXEdge/internals/providers"
)

func (c *Client) Interactive() {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Println("API key non-existant. Please do copilot-cfg <API KEY>")
		}
	}()
	cfg, err := c.configRead()
	if err != nil {
		panic(err)
	}
	engine := providers.Engine{
		Provider: c.Engine,
		Version:  c.EngineVERSION,
	}
	a := providers.Client{
		API:    cfg.Apikey,
		Engine: engine,
	}
	a.Interactive(cfg.SerialNumber)
	// c.Interactive_Telemetry()
}
