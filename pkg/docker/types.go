package docker

import (
	"encoding/json"
	"github.com/pkg/errors"
)

func (d *dockerImpl) createContextMeta(
	name string,
	host string,
) ContextMeta {
	return ContextMeta{
		Name:     name,
		Metadata: struct{}{},
		Endpoints: ContextEndpoints{
			Docker: ContextEndpoint{
				Host:          host,
				SkipTLSVerify: false,
			},
		},
	}
}

type ContextMeta struct {
	Name      string           `json:"Name"`
	Metadata  struct{}         `json:"Metadata"`
	Endpoints ContextEndpoints `json:"Endpoints"`
}

func (c *ContextMeta) JSON() ([]byte, error){
	data, err := json.Marshal(c)

	if err != nil {
		return nil, errors.Wrap(err, "failed to convert context meta to JSON")
	}

	return data, nil
}

type ContextEndpoints struct {
	Docker ContextEndpoint `json:"docker"`
}

type ContextEndpoint struct {
	Host          string `json:"Host"`
	SkipTLSVerify bool   `json:"SkipTLSVerify"`
}
