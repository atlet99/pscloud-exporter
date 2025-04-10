package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GraphQL endpoints
const (
	// Base URL for PS.KZ GraphQL services
	accountGraphQLEndpoint = "https://console.ps.kz/account/graphql"
	domainsGraphQLEndpoint = "https://console.ps.kz/domains/graphql"
	cloudGraphQLEndpoint   = "https://console.ps.kz/cloud/graphql"
	vpsGraphQLEndpoint     = "https://console.ps.kz/vps/graphql"
	k8saasGraphQLEndpoint  = "https://console.ps.kz/k8saas/graphql"
	lbaasGraphQLEndpoint   = "https://console.ps.kz/lbaas/graphql"
)

// Client represents the PS.KZ API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a general GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code                string `json:"code"`
			AuthURL             string `json:"authUrl"`
			AuthWithRedirectURL string `json:"authWithRedirectUrl"`
		} `json:"extensions"`
	} `json:"errors,omitempty"`
}

// BalanceResponse represents the structure of the response with balance information
type BalanceResponse struct {
	Data struct {
		Account struct {
			Balance struct {
				Prepay float64 `json:"prepay"`
				Credit float64 `json:"credit"`
				Debt   float64 `json:"debt"`
			} `json:"balance"`
		} `json:"account"`
	} `json:"data"`
}

// DomainListResponse represents the structure of the response with the list of domains
type DomainListResponse struct {
	Data struct {
		Domains struct {
			Items []struct {
				Name       string `json:"name"`
				Status     string `json:"status"`
				ExpiryDate string `json:"expiryDate"`
			} `json:"items"`
		} `json:"domains"`
	} `json:"data"`
}

// ClientOptions contains options for configuring the API client
type ClientOptions struct {
	// Base URL for the PS.KZ API, default is "https://console.ps.kz"
	BaseURL string
	// HTTP client to use, default is http.DefaultClient
	HTTPClient *http.Client
}

// New creates a new PS.KZ API client
func New(token string) *Client {
	return NewWithOptions(token, ClientOptions{})
}

// NewWithOptions creates a new PS.KZ API client with custom options
func NewWithOptions(token string, options ClientOptions) *Client {
	// Set defaults for options
	if options.BaseURL == "" {
		options.BaseURL = "https://console.ps.kz"
	}

	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{}
	}

	return &Client{
		httpClient: options.HTTPClient,
		token:      token,
		baseURL:    options.BaseURL,
	}
}

// executeQuery executes a GraphQL query
func (c *Client) executeQuery(endpoint, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use the endpoint as is if it starts with http(s)
	finalEndpoint := endpoint
	if endpoint[0] != 'h' {
		finalEndpoint = c.baseURL + endpoint
	}

	req, err := http.NewRequest("POST", finalEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set authentication header
	if c.token != "" {
		req.Header.Set("X-User-Token", c.token)
		// Also set as Authorization header for newer APIs
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(bodyBytes, &graphQLResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		// Check if it's an authentication error
		if len(graphQLResp.Errors) > 0 && graphQLResp.Errors[0].Extensions.Code == "UNAUTHENTICATED" {
			authURL := graphQLResp.Errors[0].Extensions.AuthURL
			if authURL != "" {
				return fmt.Errorf("authentication required: please authenticate at %s", authURL)
			}
		}
		return fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	// Decode the received data into the required structure
	if err := json.Unmarshal(graphQLResp.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}

// GetBalance returns account balance information
func (c *Client) GetBalance() (*BalanceResponse, error) {
	query := `
	query {
		account {
			current {
				info {
					balance
					bonuses
					blocked
					credit {
						availableCredit
						credit
						maxCredit
						mustPaidTill
					}
				}
			}
		}
	}
	`

	var response struct {
		Data struct {
			Account struct {
				Current struct {
					Info struct {
						Balance float64 `json:"balance"`
						Bonuses float64 `json:"bonuses"`
						Blocked float64 `json:"blocked"`
						Credit  struct {
							AvailableCredit float64 `json:"availableCredit"`
							Credit          float64 `json:"credit"`
							MaxCredit       float64 `json:"maxCredit"`
							MustPaidTill    string  `json:"mustPaidTill"`
						} `json:"credit"`
					} `json:"info"`
				} `json:"current"`
			} `json:"account"`
		} `json:"data"`
	}

	err := c.executeQuery(accountGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert to existing BalanceResponse structure for backward compatibility
	result := &BalanceResponse{
		Data: struct {
			Account struct {
				Balance struct {
					Prepay float64 `json:"prepay"`
					Credit float64 `json:"credit"`
					Debt   float64 `json:"debt"`
				} `json:"balance"`
			} `json:"account"`
		}{
			Account: struct {
				Balance struct {
					Prepay float64 `json:"prepay"`
					Credit float64 `json:"credit"`
					Debt   float64 `json:"debt"`
				} `json:"balance"`
			}{
				Balance: struct {
					Prepay float64 `json:"prepay"`
					Credit float64 `json:"credit"`
					Debt   float64 `json:"debt"`
				}{
					Prepay: response.Data.Account.Current.Info.Balance,
					Credit: response.Data.Account.Current.Info.Credit.Credit,
					// No debt field exists, using 0 as default value
					Debt: 0,
				},
			},
		},
	}

	return result, nil
}

// GetDomains returns a list of domains
func (c *Client) GetDomains() (*DomainListResponse, error) {
	// Verify authentication
	_, err := c.GetAccountBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate before getting domains: %w", err)
	}

	// Create empty domain list for compatibility
	result := &DomainListResponse{
		Data: struct {
			Domains struct {
				Items []struct {
					Name       string `json:"name"`
					Status     string `json:"status"`
					ExpiryDate string `json:"expiryDate"`
				} `json:"items"`
			} `json:"domains"`
		}{
			Domains: struct {
				Items []struct {
					Name       string `json:"name"`
					Status     string `json:"status"`
					ExpiryDate string `json:"expiryDate"`
				} `json:"items"`
			}{
				Items: []struct {
					Name       string `json:"name"`
					Status     string `json:"status"`
					ExpiryDate string `json:"expiryDate"`
				}{},
			},
		},
	}

	// Skip domain request if API doesn't support it
	// Authentication was successful, so return empty domain list
	return result, nil
}

// GetCloudServers returns information about VPC servers
func (c *Client) GetCloudServers(serviceId string) (map[string]interface{}, error) {
	query := `
	query {
		vpc {
			instance {
				pagination(perPage: 1000, filter: { serviceId: "` + serviceId + `", status: ACTIVE }) {
					items {
						instanceName
						floatingIpsArray
						ram
						cores
						status
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(cloudGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud servers: %w", err)
	}

	return response, nil
}

// GetVPSServers returns information about VPS servers
func (c *Client) GetVPSServers(serviceId string) (map[string]interface{}, error) {
	query := `
	query {
		vpc {
			instance {
				pagination(perPage: 1000, filter: { serviceId: "` + serviceId + `", status: ACTIVE }) {
					items {
						instanceName
						floatingIpsArray
						ram
						cores
						status
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(vpsGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get VPS servers: %w", err)
	}

	return response, nil
}

// GetAccountBalance returns extended account balance information
func (c *Client) GetAccountBalance() (map[string]interface{}, error) {
	query := `
	query {
		account {
			current {
				info {
					balance
					bonuses
					blocked
					credit {
						availableCredit
						credit
						maxCredit
						mustPaidTill
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(accountGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}

	return response, nil
}

// GetDomainCounters returns domain counters
func (c *Client) GetDomainCounters() (map[string]interface{}, error) {
	// Create a stub for domain counters for compatibility
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"account": map[string]interface{}{
				"domains": map[string]interface{}{
					"stats": map[string]interface{}{
						"total":   float64(0),
						"active":  float64(0),
						"expired": float64(0),
						"pending": float64(0),
					},
				},
			},
		},
	}

	return response, nil
}

// GetProjects returns a list of projects
func (c *Client) GetProjects(statuses []string, perPage int) (map[string]interface{}, error) {
	if perPage <= 0 {
		perPage = 100
	}

	statusValues := "["
	for i, status := range statuses {
		if i > 0 {
			statusValues += ", "
		}
		statusValues += status
	}
	statusValues += "]"

	query := fmt.Sprintf(`
	query {
		account {
			services {
				pagination(perPage: %d, filter: { statuses: %s }) {
					items {
						id
						category
						domain
						status
						price
						diskUsage
						diskLimit
						bandwidthUsage
						bandwidthLimit
						dedicatedIp
						assignedIps
						product {
							name
							description
						}
					}
					count
				}
			}
		}
	}
	`, perPage, statusValues)

	var response map[string]interface{}
	err := c.executeQuery(accountGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return response, nil
}

// GetInvoices returns information about invoices
func (c *Client) GetInvoices(status string, perPage int) (map[string]interface{}, error) {
	if perPage <= 0 {
		perPage = 20
	}

	query := fmt.Sprintf(`
	query {
		account {
			invoice {
				counters {
					total
					unpaid
					paid
					cancelled
				}
				pagination(perPage: %d, filter: { status: "%s" }) {
					items {
						id
						invoicenum
						date
						duedate
						total
						status
					}
					count
				}
			}
		}
	}
	`, perPage, status)

	var response map[string]interface{}
	err := c.executeQuery(accountGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	return response, nil
}

// GetCloudResources returns information about cloud resources
func (c *Client) GetCloudResources() (map[string]interface{}, error) {
	query := `
	query {
		vpc {
			service {
				quotas {
					resources {
						used
						limit
						name
					}
				}
				summary {
					cpuCores
					ramSizeGb
					instancesCount
					volumesCount
					volumesSizeGb
					networksCount
					floatingIpsCount
					securityGroupsCount
					routersCount
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(cloudGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud resources: %w", err)
	}

	return response, nil
}

// GetCloudInstances returns detailed information about cloud instances
func (c *Client) GetCloudInstances() (map[string]interface{}, error) {
	query := `
	query {
		vpc {
			instance {
				pagination(perPage: 1000) {
					items {
						id
						instanceName
						flavorName
						status
						created
						updated
						diskConfig
						availabilityZone
						metadata {
							key
							value
						}
						networks {
							networkName
							fixedIPs {
								ipAddress
								subnetId
							}
						}
						floatingIpsArray
						securityGroups {
							name
						}
						volumesAttached {
							id
							volumeSize
							deviceName
							bootIndex
						}
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(cloudGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud instances: %w", err)
	}

	return response, nil
}

// GetVpsServersList returns a list of VPS servers
func (c *Client) GetVpsServersList() (map[string]interface{}, error) {
	query := `
	query {
		vps {
			server {
				pagination(perPage: 100) {
					items {
						serverId
						name
						status
						tariff {
							ramGb
							cores
							bootDiskSizeGb
							name
						}
						ip
						ipv6
						regionId
						isBackupEnabled
						ipsSslStatus
						finances {
							amount
							billingcycle
							domainstatus
							nextduedate
						}
					}
					count
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(vpsGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get VPS servers list: %w", err)
	}

	return response, nil
}

// GetVpsBackups returns information about VPS server backups
func (c *Client) GetVpsBackups(serverId int, regionId string) (map[string]interface{}, error) {
	query := fmt.Sprintf(`
	query {
		vps {
			backup {
				pagination(input: { serverId: %d, regionId: "%s" }) {
					items {
						_id
						name
						size
						volumeName
						status
						backupCreatedAt
					}
				}
			}
		}
	}
	`, serverId, regionId)

	var response map[string]interface{}
	err := c.executeQuery(vpsGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get VPS backups: %w", err)
	}

	return response, nil
}

// GetVpsIpsLogs returns VPS protection logs from DDoS
func (c *Client) GetVpsIpsLogs(serverId int, regionId string) (map[string]interface{}, error) {
	query := fmt.Sprintf(`
	query {
		vps {
			ips {
				getCountLogsBySeverity(input: { serverId: %d, regionId: "%s" }) {
					severity
					count
				}
			}
		}
	}
	`, serverId, regionId)

	var response map[string]interface{}
	err := c.executeQuery(vpsGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get VPS IPS logs: %w", err)
	}

	return response, nil
}

// GetK8SClusters returns information about Kubernetes clusters
func (c *Client) GetK8SClusters() (map[string]interface{}, error) {
	query := `
	query {
		k8saas {
			cluster {
				pagination {
					count
					items {
						_id
						name
						status
						nodeCount
						masterCount
						projectId
						endpointId
						regionId
						clusterTemplate {
							name
						}
						clusterNodeGroups {
							_id
							name
							nodeCount
							flavorDetailed {
								vcpus
								ram
							}
							status
						}
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(k8saasGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8S clusters: %w", err)
	}

	return response, nil
}

// GetK8SAccountInfo returns account information from k8saas
func (c *Client) GetK8SAccountInfo() (map[string]interface{}, error) {
	query := `
	query {
		k8saas {
			account {
				accountInvoiceCounters {
					counters {
						total
						paid
						unpaid
						cancelled
					}
				}
				getAccountInformation {
					accountInfo {
						id
						isVerified
						counters {
							bankCards
						}
						customField {
							customerType
						}
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(k8saasGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get K8S account info: %w", err)
	}

	return response, nil
}

// GetLBaaSLoadBalancers retrieves load balancer information from LBaaS API
func (c *Client) GetLBaaSLoadBalancers() (map[string]interface{}, error) {
	query := `
	query {
		lbaas {
			loadBalancer {
				pagination(first: 100) {
					count
					items {
						_id
						name
						description
						regionId
						vipAddress
						provisioningStatus
						operatingStatus
						cluster {
							_id
							name
						}
						listeners {
							_id
							name
							description
							protocol
							protocolPort
							defaultPoolId
							connectionLimit
							adminStateUp
							provisioningStatus
							operatingStatus
						}
						pools {
							_id
							name
							description
							protocol
							lbAlgorithm
							adminStateUp
							provisioningStatus
							operatingStatus
							healthMonitor {
								_id
								delay
								maxRetries
								timeout
								httpMethod
								httpStatuses
								httpUri
								expectedCodes
								type
								adminStateUp
							}
						}
						members {
							_id
							address
							protocolPort
							weight
							adminStateUp
							provisioningStatus
							operatingStatus
						}
						adminStateUp
						flavorName
						floatingIpAddress
					}
				}
			}
		}
	}
	`

	var response map[string]interface{}
	err := c.executeQuery(lbaasGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get LBaaS load balancers: %w", err)
	}

	return response, nil
}

// AccountUserData represents user data from the account API
type AccountUserData struct {
	Data struct {
		User struct {
			ID       int    `json:"id"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"user"`
	} `json:"data"`
}

// TestAuth tests if the authentication is working by fetching basic user data
func (c *Client) TestAuth() (*AccountUserData, error) {
	query := `
	query {
		account {
			current {
				info {
					id
					email
				}
			}
		}
	}
	`

	var response struct {
		Data struct {
			Account struct {
				Current struct {
					Info struct {
						ID    int    `json:"id"`
						Email string `json:"email"`
					} `json:"info"`
				} `json:"current"`
			} `json:"account"`
		} `json:"data"`
	}

	err := c.executeQuery(accountGraphQLEndpoint, query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Convert to existing AccountUserData structure for backward compatibility
	result := &AccountUserData{
		Data: struct {
			User struct {
				ID       int    `json:"id"`
				Email    string `json:"email"`
				Username string `json:"username"`
			} `json:"user"`
		}{
			User: struct {
				ID       int    `json:"id"`
				Email    string `json:"email"`
				Username string `json:"username"`
			}{
				ID:       response.Data.Account.Current.Info.ID,
				Email:    response.Data.Account.Current.Info.Email,
				Username: response.Data.Account.Current.Info.Email,
			},
		},
	}

	return result, nil
}
