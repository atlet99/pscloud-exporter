package client

import (
	"pscloud-exporter/internal/models"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	apiURL string
	user   string
	pass   string
	rest   *resty.Client
}

func New(user, pass string) *Client {
	return &Client{
		apiURL: "https://api.ps.kz",
		user:   user,
		pass:   pass,
		rest:   resty.New(),
	}
}

func (c *Client) GetBalance() (*models.BalanceResponse, error) {
	var resp models.BalanceResponse
	_, err := c.rest.R().
		SetQueryParams(map[string]string{
			"username":      c.user,
			"password":      c.pass,
			"input_format":  "http",
			"output_format": "json",
		}).
		SetResult(&resp).
		Get(c.apiURL + "/client/get-balance")
	return &resp, err
}
