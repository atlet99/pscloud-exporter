package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// API endpoints
const (
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

	// OIDC endpoints
	oidcTokenEndpoint = "https://auth.ps.kz/oidc/token"
)

// Client represents PS.KZ API client
// Note: API credentials (username and password) are different from your personal account credentials.
// To get API access, visit: https://old.ps.kz/client/api
type Client struct {
	client       *resty.Client
	username     string
	password     string
	baseURL      string
	token        *Token
	useHTTP      bool
	clientSecret string
	codeVerifier string
}

// Token represents OIDC token response
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	ExpiresAt    time.Time
}

// Response represents common API response structure
type Response struct {
	Result string      `json:"result"`
	Answer interface{} `json:"answer"`
	Body   []byte      // Raw response body
}

// BalanceResponse represents balance information response
type BalanceResponse struct {
	Answer struct {
		Prepay     string `json:"prepay"`
		Credit     string `json:"credit"`
		CreditInfo struct {
			Debt string `json:"debt"`
		} `json:"credit_info"`
	} `json:"answer"`
}

// DomainListResponse represents domain list response
type DomainListResponse struct {
	Result string   `json:"result"`
	Answer []Domain `json:"answer"`
}

// Domain represents domain information
type Domain struct {
	Domain     string `json:"domain"`
	ExpiryDate string `json:"expirydate"`
	Status     string `json:"status"`
}

// ClientDomainListResponse represents client domain list response
type ClientDomainListResponse struct {
	Result string   `json:"result"`
	Answer []Domain `json:"answer"`
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

// generateCodeVerifier generates a random code verifier for PKCE
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates a code challenge from the code verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// New creates a new PS.KZ API client
func New(username, password, baseURL string, useHTTP bool) *Client {
	c := &Client{
		client:   resty.New(),
		username: username,
		password: password,
		baseURL:  baseURL,
		useHTTP:  useHTTP,
	}

	// Generate code_verifier for PKCE
	verifier, err := generateCodeVerifier()
	if err != nil {
		// Use default value if generation fails
		verifier = "default_code_verifier"
	}
	c.codeVerifier = verifier

	// Use username as client_secret
	c.clientSecret = username

	return c
}

// SetBaseURL sets the base URL for API requests
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// authenticate performs OIDC authentication and obtains a token
func (c *Client) authenticate() error {
	challenge := generateCodeChallenge(c.codeVerifier)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":            "password",
			"username":              c.username,
			"password":              c.password,
			"client_id":             "ps.kz",
			"client_secret":         c.clientSecret,
			"code_verifier":         c.codeVerifier,
			"code_challenge":        challenge,
			"code_challenge_method": "S256",
		}).
		Post(oidcTokenEndpoint)

	if err != nil {
		return fmt.Errorf("authentication request failed: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	var token Token
	if err := json.Unmarshal(resp.Body(), &token); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	c.token = &token

	// Set token for all subsequent requests
	c.client.SetAuthToken(token.AccessToken)

	return nil
}

// ensureToken ensures that we have a valid token
func (c *Client) ensureToken() error {
	if c.token == nil || time.Now().After(c.token.ExpiresAt) {
		return c.authenticate()
	}
	return nil
}

// Do executes API request with authentication
func (c *Client) Do(method, path string) (*Response, error) {
	// Ensure we have a valid token
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	client := resty.New().
		SetBaseURL(c.baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+c.token.AccessToken).
		SetQueryParams(map[string]string{
			"output_format": outputFormatJSON,
		})

	resp, err := client.R().Execute(method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var response Response
	response.Body = resp.Body() // Save raw response body
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetBalance returns account balance information
func (c *Client) GetBalance() (*BalanceResponse, error) {
	resp, err := c.Do("GET", balanceEndpoint)
	if err != nil {
		return nil, err
	}

	var balanceResp BalanceResponse
	if err := json.Unmarshal(resp.Body, &balanceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balance response: %w", err)
	}

	return &balanceResp, nil
}

// GetClientDomainList returns list of client domains
func (c *Client) GetClientDomainList() (*ClientDomainListResponse, error) {
	resp, err := c.Do("GET", clientDomainListEndpoint)
	if err != nil {
		return nil, err
	}

	var domainResp ClientDomainListResponse
	if err := json.Unmarshal(resp.Body, &domainResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client domain list response: %w", err)
	}

	return &domainResp, nil
}

// GetDomainList returns list of domains
func (c *Client) GetDomainList() (*DomainListResponse, error) {
	resp, err := c.Do("GET", domainListEndpoint)
	if err != nil {
		return nil, err
	}

	var domainResp DomainListResponse
	if err := json.Unmarshal(resp.Body, &domainResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal domain list response: %w", err)
	}

	return &domainResp, nil
}

// GetDomainNSSList returns list of domain nameservers
func (c *Client) GetDomainNSSList(domain string) (*DomainNSSListResponse, error) {
	resp, err := c.Do("GET", domainNSSListEndpoint)
	if err != nil {
		return nil, err
	}

	var nssResp DomainNSSListResponse
	if err := json.Unmarshal(resp.Body, &nssResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal domain NSS list response: %w", err)
	}

	return &nssResp, nil
}

// GetProfileData returns profile data
func (c *Client) GetProfileData() (*ProfileResponse, error) {
	resp, err := c.Do("GET", profileDataEndpoint)
	if err != nil {
		return nil, err
	}

	var profileResp ProfileResponse
	if err := json.Unmarshal(resp.Body, &profileResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile data response: %w", err)
	}

	return &profileResp, nil
}

// GetProductList returns list of products
func (c *Client) GetProductList() (*ProductListResponse, error) {
	resp, err := c.Do("GET", productListEndpoint)
	if err != nil {
		return nil, err
	}

	var productResp ProductListResponse
	if err := json.Unmarshal(resp.Body, &productResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product list response: %w", err)
	}

	return &productResp, nil
}

// GetProductDetails returns product details
func (c *Client) GetProductDetails(productID string) (*ProductListResponse, error) {
	resp, err := c.Do("GET", productDetailsEndpoint)
	if err != nil {
		return nil, err
	}

	var productResp ProductListResponse
	if err := json.Unmarshal(resp.Body, &productResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product details response: %w", err)
	}

	return &productResp, nil
}

// GetNSS returns DNS servers for the specified domain
func (c *Client) GetNSS(domain string) (*NSSResponse, error) {
	resp, err := c.Do("GET", nssGetEndpoint)
	if err != nil {
		return nil, err
	}

	var nssResp NSSResponse
	if err := json.Unmarshal(resp.Body, &nssResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NSS response: %w", err)
	}

	return &nssResp, nil
}

// GetContactWhois returns contact information by handle
func (c *Client) GetContactWhois(handle string) (*ContactWhoisResponse, error) {
	resp, err := c.Do("GET", contactWhoisEndpoint)
	if err != nil {
		return nil, err
	}

	var contactResp ContactWhoisResponse
	if err := json.Unmarshal(resp.Body, &contactResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contact whois response: %w", err)
	}

	return &contactResp, nil
}

// GetTransferPassword returns transfer password for the specified domain
func (c *Client) GetTransferPassword(domain string) (*PasswordGetResponse, error) {
	resp, err := c.Do("GET", passwordGetEndpoint)
	if err != nil {
		return nil, err
	}

	var passwordResp PasswordGetResponse
	if err := json.Unmarshal(resp.Body, &passwordResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal password get response: %w", err)
	}

	return &passwordResp, nil
}

// GetNSInfo returns information about a nameserver in .kz zone
func (c *Client) GetNSInfo(host string) (*NSInfoResponse, error) {
	resp, err := c.Do("GET", nsInfoEndpoint)
	if err != nil {
		return nil, err
	}

	var nsResp NSInfoResponse
	if err := json.Unmarshal(resp.Body, &nsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NS info response: %w", err)
	}

	return &nsResp, nil
}

// GetInvoiceDetails returns invoice details
func (c *Client) GetInvoiceDetails(invoiceID string) (*InvoiceDetailsResponse, error) {
	resp, err := c.Do("GET", invoiceDetailsEndpoint)
	if err != nil {
		return nil, err
	}

	var invoiceResp InvoiceDetailsResponse
	if err := json.Unmarshal(resp.Body, &invoiceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invoice details response: %w", err)
	}

	return &invoiceResp, nil
}

// GetUsername returns the username used for authentication
func (c *Client) GetUsername() string {
	return c.username
}
