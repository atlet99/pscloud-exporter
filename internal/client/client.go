package client

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "https://api.ps.kz"

	// API endpoints
	balanceEndpoint          = "/client/get-balance"
	clientDomainListEndpoint = "/client/get-domain-list" // Client API endpoint
	domainListEndpoint       = "/domain/domain-list"     // Domain API endpoint
	domainNSSListEndpoint    = "/domain/domain-nss-list"
	profileDataEndpoint      = "/client/get-profile-data"
	productListEndpoint      = "/client/get-product-list"
	productDetailsEndpoint   = "/client/get-product-details"
	nssGetEndpoint           = "/client/nss-get"
	contactWhoisEndpoint     = "/client/contact-whois"
	passwordGetEndpoint      = "/client/password-get"
	nsInfoEndpoint           = "/client/ns-info"
	invoiceDetailsEndpoint   = "/invoice/invoice-details"

	// Output format
	outputFormatJSON = "json"
)

// Client represents PS.KZ API client
type Client struct {
	client   *resty.Client
	username string
	password string
}

// Common response structures
type Response struct {
	Result string      `json:"result"`
	Answer interface{} `json:"answer"`
}

// Balance related structures
type BalanceResponse struct {
	Result string      `json:"result"`
	Answer BalanceInfo `json:"answer"`
}

type BalanceInfo struct {
	Prepay     string     `json:"prepay"`
	Credit     string     `json:"credit"`
	CreditInfo CreditInfo `json:"credit_info"`
}

type CreditInfo struct {
	Active  string `json:"active"`
	Debt    string `json:"debt"`
	PayTill string `json:"pay_till"`
}

// Domain related structures
type DomainListResponse struct {
	Result string   `json:"result"`
	Answer []Domain `json:"answer"`
}

type Domain struct {
	Domain     string `json:"domain"`
	ExpiryDate string `json:"expirydate"`
	Status     string `json:"status"`
}

// Profile related structures
type ProfileResponse struct {
	Result string    `json:"result"`
	Answer []Profile `json:"answer"`
}

type Profile struct {
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	CompanyName string `json:"companyname"`
	Email       string `json:"email"`
	City        string `json:"city"`
	State       string `json:"state"`
	Address     string `json:"adress"`
	PhoneNumber string `json:"phonenumber"`
}

// Product related structures
type ProductListResponse struct {
	Result string        `json:"result"`
	Answer ProductAnswer `json:"answer"`
}

type ProductAnswer struct {
	Products Product `json:"products"`
}

type Product struct {
	Description     string         `json:"description"`
	Domain          string         `json:"domain"`
	RegDate         string         `json:"regdate"`
	NextInvoiceDate string         `json:"nextinvoicedate"`
	BillingCycle    string         `json:"billingcycle"`
	Status          string         `json:"status"`
	Amount          string         `json:"amount"`
	Options         ProductOptions `json:"options"`
}

type ProductOptions struct {
	DedicatedIP string `json:"dedicatedip"`
	DiskUsage   string `json:"diskusage"`
	DiskLimit   string `json:"disklimit"`
	BWUsage     string `json:"bwusage"`
	BWLimit     string `json:"bwlimit"`
}

// NSS related structures
type NSSResponse struct {
	Result string   `json:"result"`
	Answer []string `json:"answer"`
}

// Contact Whois related structures
type ContactWhoisResponse struct {
	Result string       `json:"result"`
	Answer ContactWhois `json:"answer"`
}

type ContactWhois struct {
	Handle  string `json:"handle"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	City    string `json:"city"`
	Country string `json:"country"`
	Type    string `json:"type"`
}

// Password Get related structures
type PasswordGetResponse struct {
	Result string `json:"result"`
	Answer string `json:"answer"`
}

// NS Info related structures
type NSInfoResponse struct {
	Result string `json:"result"`
	Answer NSInfo `json:"answer"`
}

type NSInfo struct {
	Host   string   `json:"host"`
	IPs    []string `json:"ips"`
	Status string   `json:"status"`
}

// Domain NSS List related structures
type DomainNSSListResponse struct {
	Result string         `json:"result"`
	Answer NameserverInfo `json:"answer"`
}

type NameserverInfo struct {
	Nameservers struct {
		NS []string `json:"ns"`
	} `json:"nameservers"`
}

// Invoice related structures
type InvoiceDetailsResponse struct {
	Result string         `json:"result"`
	Answer InvoiceDetails `json:"answer"`
}

type InvoiceDetails struct {
	ID            string        `json:"id"`
	Date          string        `json:"date"`
	DatePaid      string        `json:"datepaid"`
	PaymentMethod string        `json:"paymentmethod"`
	Status        string        `json:"status"`
	Total         string        `json:"total"`
	Credit        string        `json:"credit"`
	Items         []InvoiceItem `json:"items"`
}

type InvoiceItem struct {
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

// New creates a new PS.KZ API client
func New(username, password string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"output_format": outputFormatJSON,
		})

	return &Client{
		client:   client,
		username: username,
		password: password,
	}
}

// GetBalance returns account balance information
func (c *Client) GetBalance() (*BalanceResponse, error) {
	var response BalanceResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
		}).
		SetResult(&response).
		Get(balanceEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetClientDomainList returns list of domains using client API
func (c *Client) GetClientDomainList() (*DomainListResponse, error) {
	var response DomainListResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username":      c.username,
			"password":      c.password,
			"input_format":  "http",
			"output_format": outputFormatJSON,
		}).
		SetResult(&response).
		Get(clientDomainListEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get client domain list: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetDomainList returns list of domains using domain API
func (c *Client) GetDomainList() (*DomainListResponse, error) {
	var response DomainListResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username":      c.username,
			"password":      c.password,
			"input_format":  "http",
			"output_format": outputFormatJSON,
		}).
		SetResult(&response).
		Get(domainListEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get domain list: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetDomainNSSList returns DNS servers for the specified domain
func (c *Client) GetDomainNSSList(domain string) (*DomainNSSListResponse, error) {
	var response DomainNSSListResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username":      c.username,
			"password":      c.password,
			"input_format":  "http",
			"output_format": outputFormatJSON,
			"dname":         domain,
		}).
		SetResult(&response).
		Get(domainNSSListEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get domain NSS list: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetProfileData returns profile information
func (c *Client) GetProfileData() (*ProfileResponse, error) {
	var response ProfileResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
		}).
		SetResult(&response).
		Get(profileDataEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get profile data: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetProductList returns list of products/services
func (c *Client) GetProductList() (*ProductListResponse, error) {
	var response ProductListResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
		}).
		SetResult(&response).
		Get(productListEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get product list: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetProductDetails returns details of a specific product/service
func (c *Client) GetProductDetails(productID string) (*ProductListResponse, error) {
	var response ProductListResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username":  c.username,
			"password":  c.password,
			"productId": productID,
		}).
		SetResult(&response).
		Get(productDetailsEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get product details: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetNSS returns DNS servers for the specified domain
func (c *Client) GetNSS(domain string) (*NSSResponse, error) {
	var response NSSResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
			"domain":   domain,
		}).
		SetResult(&response).
		Get(nssGetEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get NSS: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetContactWhois returns contact information by handle
func (c *Client) GetContactWhois(handle string) (*ContactWhoisResponse, error) {
	var response ContactWhoisResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
			"handle":   handle,
		}).
		SetResult(&response).
		Get(contactWhoisEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get contact whois: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetTransferPassword returns transfer password for the specified domain
func (c *Client) GetTransferPassword(domain string) (*PasswordGetResponse, error) {
	var response PasswordGetResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
			"domain":   domain,
		}).
		SetResult(&response).
		Get(passwordGetEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get transfer password: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetNSInfo returns information about a nameserver in .kz zone
func (c *Client) GetNSInfo(host string) (*NSInfoResponse, error) {
	var response NSInfoResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username": c.username,
			"password": c.password,
			"host":     host,
		}).
		SetResult(&response).
		Get(nsInfoEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get NS info: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}

// GetInvoiceDetails returns details of a specific invoice
func (c *Client) GetInvoiceDetails(invoiceID string) (*InvoiceDetailsResponse, error) {
	var response InvoiceDetailsResponse

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"username":      c.username,
			"password":      c.password,
			"input_format":  "http",
			"output_format": outputFormatJSON,
			"invoiceId":     invoiceID,
		}).
		SetResult(&response).
		Get(invoiceDetailsEndpoint)

	if err != nil {
		return nil, fmt.Errorf("failed to get invoice details: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if response.Result != "success" {
		return nil, fmt.Errorf("API error: %s", response.Result)
	}

	return &response, nil
}
