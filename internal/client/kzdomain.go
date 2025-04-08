package client

import (
	"fmt"
)

const (
	// KZ Domain API endpoints - only read operations
	domainCheckEndpoint = "/kzdomain/domain-check"
	domainWhoisEndpoint = "/kzdomain/domain-whois"
	getPricesEndpoint   = "/kzdomain/get-prices"
)

// Domain check structures
type DomainCheckResponse struct {
	Result string          `json:"result"`
	Answer DomainCheckInfo `json:"answer"`
}

type DomainCheckInfo struct {
	Domains []DomainCheckResult `json:"domains"`
}

type DomainCheckResult struct {
	DName     string `json:"dname"`
	Result    string `json:"result"`
	ErrorCode string `json:"error_code,omitempty"`
	ErrorText string `json:"error_text,omitempty"`
}

// Domain whois structures
type DomainWhoisResponse struct {
	Result string          `json:"result"`
	Answer DomainWhoisInfo `json:"answer"`
}

type DomainWhoisInfo struct {
	DName             string         `json:"dname"`
	Registrar         string         `json:"registrar"`
	RegistrantHandle  string         `json:"registrant_handle"`
	AdminHandle       string         `json:"admin_handle"`
	RegistrantContact *ContactInfo   `json:"registrant_contact,omitempty"`
	AdminContact      *ContactInfo   `json:"admin_contact,omitempty"`
	Nameservers       NameserverList `json:"nameservers"`
	Statuses          StatusList     `json:"statuses"`
	Expire            TimestampInfo  `json:"expire"`
	Create            RegistrarInfo  `json:"create"`
	Update            *RegistrarInfo `json:"update,omitempty"`
	Transfer          *TimestampInfo `json:"transfer,omitempty"`
	SrvLoc            ServerLocation `json:"srvloc"`
}

type ContactInfo struct {
	Handle    string         `json:"handle"`
	Name      string         `json:"name"`
	Org       string         `json:"org"`
	Street    string         `json:"street"`
	City      string         `json:"city"`
	State     string         `json:"sp"`
	PostCode  string         `json:"pc"`
	Country   string         `json:"cc"`
	Phone     string         `json:"voice"`
	PhoneExt  string         `json:"voiceext"`
	Fax       string         `json:"fax"`
	FaxExt    string         `json:"faxext"`
	Email     string         `json:"email"`
	Registrar string         `json:"registrar"`
	Create    RegistrarInfo  `json:"create"`
	Update    *RegistrarInfo `json:"update,omitempty"`
}

type NameserverList struct {
	NS []string `json:"ns"`
}

type StatusList struct {
	Status []string `json:"status"`
}

type TimestampInfo struct {
	UTC  string `json:"utc"`
	Unix string `json:"unix"`
}

type RegistrarInfo struct {
	ID   string `json:"id"`
	UTC  string `json:"utc"`
	Unix string `json:"unix"`
}

type ServerLocation struct {
	State  string `json:"sp"`
	City   string `json:"city"`
	Street string `json:"street"`
}

// Domain check method - safe, read-only operation
func (c *Client) DomainCheck(domains []string) (*DomainCheckResponse, error) {
	var response DomainCheckResponse

	params := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	for i, domain := range domains {
		params[fmt.Sprintf("dname[%d]", i)] = domain
	}

	resp, err := c.client.R().
		SetQueryParams(params).
		SetResult(&response).
		Get(domainCheckEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to check domains: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// Domain whois method - safe, read-only operation
func (c *Client) DomainWhois(domain string, contactWhois bool) (*DomainWhoisResponse, error) {
	var response DomainWhoisResponse

	params := map[string]string{
		"username": c.username,
		"password": c.password,
		"dname":    domain,
	}

	if contactWhois {
		params["contact_whois"] = "1"
	}

	resp, err := c.client.R().
		SetQueryParams(params).
		SetResult(&response).
		Get(domainWhoisEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get domain whois: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetPrices method - safe, read-only operation
type PricesResponse struct {
	Result string     `json:"result"`
	Answer PricesInfo `json:"answer"`
}

type PricesInfo struct {
	ZoneKZ    ZonePrice `json:"zone_kz"`
	ZoneComKZ ZonePrice `json:"zone_com_kz"`
	ZoneOrgKZ ZonePrice `json:"zone_org_kz"`
}

type ZonePrice struct {
	Name  string      `json:"name"`
	Reg   PricePeriod `json:"reg"`
	Renew PricePeriod `json:"renew"`
}

type PricePeriod struct {
	Price     string `json:"price"`
	MinPeriod string `json:"min_period"`
	MaxPeriod string `json:"max_period"`
}

func (c *Client) GetPrices() (*PricesResponse, error) {
	var response PricesResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
		}).
		SetResult(&response).
		Get(getPricesEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}
