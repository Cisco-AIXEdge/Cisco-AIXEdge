package internals

type Client struct {
	Version       string
	SoftwareURL   string
	EngineVERSION string
	Engine        string
}

func (c *Client) Init() {
	c.Version = VERSION
	c.SoftwareURL = SOFTWARE_URL
	c.EngineVERSION = ENGINE_VERSION
	c.Engine = ENGINE
}
