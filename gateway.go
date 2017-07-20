package discgo

type Game struct {
	Name string  `json:"name"`
	Type *int    `json:"type"`
	URL  *string `json:"url"`
}

type gatewayEndpoint struct {
	*endpoint
}

func (c *Client) gateway() gatewayEndpoint {
	e2 := c.e.appendMajor("gateway")
	return gatewayEndpoint{e2}
}

func (g gatewayEndpoint) get() (url string, err error) {
	var urlStruct struct {
		URL string `json:"url"`
	}
	return urlStruct.URL, g.doMethod("GET", nil, &urlStruct)
}
