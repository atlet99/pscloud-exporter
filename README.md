# PSCloud Exporter

![Go version](https://img.shields.io/github/go-mod/go-version/atlet99/pscloud-exporter/main?style=flat&label=go-version) [![Docker Image Version](https://img.shields.io/docker/v/zetfolder17/pscloud-exporter?label=docker%20image&sort=semver)](https://hub.docker.com/r/zetfolder17/pscloud-exporter) ![Docker Image Size](https://img.shields.io/docker/image-size/zetfolder17/pscloud-exporter/latest) [![CI](https://github.com/atlet99/pscloud-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/atlet99/pscloud-exporter/actions/workflows/ci.yml) [![GitHub contributors](https://img.shields.io/github/contributors/atlet99/pscloud-exporter)](https://github.com/atlet99/pscloud-exporter/graphs/contributors/) [![Go Report Card](https://goreportcard.com/badge/github.com/atlet99/pscloud-exporter)](https://goreportcard.com/report/github.com/atlet99/pscloud-exporter) [![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/atlet99/pscloud-exporter/badge)](https://securityscorecards.dev/viewer/?uri=github.com/atlet99/pscloud-exporter) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/atlet99/pscloud-exporter?sort=semver)

A Prometheus exporter for PS.KZ (PSCloud) services that collects metrics about your account balance, domains, and server resources via GraphQL API.

Repository: [github.com/atlet99/pscloud-exporter](https://github.com/atlet99/pscloud-exporter)

## Features

- Account balance metrics (prepay, credit, debt)
- Domain metrics (expiry dates, status)
- Cloud/VPS server metrics (RAM, CPU cores, status, IP count)
- Kubernetes clusters metrics (nodes, masters, status)
- Comprehensive error reporting
- GraphQL API integration
- Support for both .yml and .yaml configuration files
- Docker support

## Prerequisites

- Go 1.21 or higher
- PS.KZ API Token (created in PS.KZ Console)
- Docker (optional, for containerized deployment)

## Installation

### From Source

```bash
git clone https://github.com/atlet99/pscloud-exporter.git
cd pscloud-exporter
make build
```

### Using Docker

```bash
docker pull zetfolder17/pscloud-exporter:latest
```

### Prebuilt Binaries

Each release contains prebuilt binaries for:
- Linux (pscloud-exporter.linux)
- macOS (pscloud-exporter.darwin)
- Windows (pscloud-exporter.exe)

## Configuration

The configuration file `config.yml` supports the following options:

```yaml
# PSCloud Exporter Configuration
token: ""  # Can be left empty and set via PSCLOUD_TOKEN environment variable
serviceId: ""  # Service ID for VPC and VPS API requests (optional)
baseUrl: "https://console.ps.kz"  # Base URL for PS.KZ API (optional)

# Web server configuration
web:
  listenAddress: ":9116"
  metricsPrefix: "pskz"
  telemetryPath: "/metrics"
```

## Authentication

PSCloud Exporter uses a PS.KZ API token to retrieve metrics. To obtain a token:

1. Log in to your [PS.KZ Console](https://console.ps.kz)
2. Navigate to "Account Settings"
3. Go to the "API Integration" section
4. Create a new access token with the "Engineer" role
5. Copy the generated token

You can provide the token to the exporter in one of the following ways:

### 1. Using environment variables (recommended)

You can use either `PS_ACCOUNT_TOKEN` (preferred) or `PSCLOUD_TOKEN`:

```bash
# Using PS_ACCOUNT_TOKEN
export PS_ACCOUNT_TOKEN="your_access_token"
./bin/pscloud-exporter

# Alternative using PSCLOUD_TOKEN
export PSCLOUD_TOKEN="your_access_token"
./bin/pscloud-exporter
```

For Docker:
```bash
# Using PS_ACCOUNT_TOKEN
docker run -e PS_ACCOUNT_TOKEN="your_access_token" -p 9116:9116 zetfolder17/pscloud-exporter

# Alternative using PSCLOUD_TOKEN
docker run -e PSCLOUD_TOKEN="your_access_token" -p 9116:9116 zetfolder17/pscloud-exporter
```

### 2. Using configuration file

In the `config.yml` file:
```yaml
token: "your_access_token"
```

### 3. Using command line flag

```bash
./bin/pscloud-exporter -token="your_access_token"
```

## Verifying Configuration

To verify that your token works correctly, you can use the following command:

```bash
curl -X POST https://console.ps.kz/account/graphql \
  -H "Content-Type: application/json" \
  -H "X-User-Token: $PSCLOUD_TOKEN" \
  -d '{"query": "query { account { current { info { balance } } } }"}'
```

If the token is valid, you should receive a response with your account balance.

## Usage

### Running Locally

```bash
./bin/pscloud-exporter [flags]
```

Available flags:
- `-config`: Path to configuration file (default: "config.yml")
- `-listen-address`: Address to listen on for web interface and telemetry (default: ":9116")
- `-metrics-path`: Path under which to expose metrics (default: "/metrics")
- `-token`: PS.KZ API token (overrides config file)
- `-service-id`: PS.KZ service ID for cloud servers (overrides config file)
- `-base-url`: Base URL for PS.KZ API (default: "https://console.ps.kz")
- `-skip-auth-check`: Skip authentication validation on startup

### Running with Docker

```bash
docker run -d \
  -p 9116:9116 \
  -v $(pwd)/config.yml:/config.yml \
  zetfolder17/pscloud-exporter:latest
```

## Available Metrics

The exporter provides the following metrics:

```
# Account Metrics
pskz_prepay_balance{account="default"} <value>                # Current prepay balance
pskz_credit_balance{account="default"} <value>                # Current credit balance
pskz_debt_balance{account="default"} <value>                  # Current debt balance
pskz_bonus_balance{account="default"} <value>                 # Current bonus balance
pskz_blocked_balance{account="default"} <value>               # Current blocked balance

# Domain Metrics
pskz_domain_expiry_days{domain="example.com"} <value>         # Days until domain expiry
pskz_domain_status{domain="example.com",status="active"} <value>  # Domain status (1 = active, 0 = inactive)
pskz_domain_counters{domain="total"} <value>                  # Domain counter for total domains
pskz_domain_counters{domain="active"} <value>                 # Domain counter for active domains
pskz_domain_counters{domain="expired"} <value>                # Domain counter for expired domains
pskz_domain_counters{domain="pending"} <value>                # Domain counter for pending domains

# VPS and Cloud Server Metrics
pskz_server_status{id="server-id",name="server-name",status="active"} <value>  # Server status (1 = active)
pskz_server_ram_mb{id="server-id",name="server-name"} <value>  # Server RAM in MB
pskz_server_cores{id="server-id",name="server-name"} <value>   # Server CPU cores
pskz_server_ip_count{id="server-id",name="server-name"} <value> # Number of IPs associated with server

# Kubernetes Metrics
pskz_k8s_cluster_count{status="total"} <value>                # Total number of Kubernetes clusters
pskz_k8s_cluster_count{status="<status>"} <value>             # Number of Kubernetes clusters with specific status
pskz_k8s_cluster_status{cluster_id="id",name="name",status="status"} <value>  # Cluster status (1 = active)
pskz_k8s_cluster_nodes{cluster_id="id",name="name"} <value>   # Number of worker nodes in cluster
pskz_k8s_cluster_masters{cluster_id="id",name="name"} <value> # Number of master nodes in cluster
pskz_k8s_nodegroup_status{cluster_id="id",cluster_name="name",nodegroup_id="id",name="name"} <value>  # Node group status
pskz_k8s_nodegroup_nodes{cluster_id="id",cluster_name="name",nodegroup_id="id",name="name"} <value>   # Nodes in group
pskz_k8s_nodegroup_cores{cluster_id="id",cluster_name="name",nodegroup_id="id",name="name"} <value>   # Cores per node
pskz_k8s_nodegroup_ram{cluster_id="id",cluster_name="name",nodegroup_id="id",name="name"} <value>     # RAM per node (MB)

# K8S Project Metrics (Dynamic metrics based on project quota)
pskz_k8s_project_quota_<service>_<key>_limit{project_id="id",project_name="name",region_id="id"} <value>  # Quota limit
pskz_k8s_project_quota_<service>_<key>_used{project_id="id",project_name="name",region_id="id"} <value>   # Quota usage
pskz_k8s_project_status_count{status="<status>"} <value>      # Count of projects by status
pskz_k8s_project_type_count{type="<type>"} <value>            # Count of projects by type

# LBaaS Metrics
pskz_lbaas_loadbalancer_count{status="total"} <value>         # Total load balancers
pskz_lbaas_loadbalancer_count{status="<status>"} <value>      # Load balancers by status
pskz_lbaas_loadbalancer_status{id="id",name="name"} <value>   # Load balancer status (1 = active)
pskz_lbaas_listeners_count{loadbalancer_id="id"} <value>      # Number of listeners per load balancer
pskz_lbaas_pools_count{loadbalancer_id="id"} <value>          # Number of pools per load balancer
pskz_lbaas_members_count{loadbalancer_id="id"} <value>        # Number of members per load balancer
pskz_lbaas_floating_ip{loadbalancer_id="id",name="name"} <value>  # Whether load balancer has floating IP (1 = yes)

# Cloud Summary Metrics
pskz_cloud_summary{resource="cpu_cores"} <value>              # Total CPU cores in cloud
pskz_cloud_summary{resource="ram_gb"} <value>                 # Total RAM in cloud (GB)
pskz_cloud_summary{resource="instances_count"} <value>        # Total number of instances
pskz_cloud_summary{resource="volumes_count"} <value>          # Total number of volumes
pskz_cloud_summary{resource="volumes_size_gb"} <value>        # Total volume size (GB)
pskz_cloud_summary{resource="floating_ips_count"} <value>     # Total number of floating IPs
pskz_cloud_summary{resource="networks_count"} <value>         # Total number of networks
pskz_cloud_summary{resource="routers_count"} <value>          # Total number of routers
pskz_cloud_summary{resource="security_groups_count"} <value>  # Total number of security groups

# Invoice Metrics
pskz_invoice_counters{type="total"} <value>                   # Total invoices
pskz_invoice_counters{type="unpaid"} <value>                  # Unpaid invoices
pskz_invoice_counters{type="paid"} <value>                    # Paid invoices
pskz_invoice_counters{type="cancelled"} <value>               # Cancelled invoices

# Exporter Status Metrics
pskz_scrape_duration_seconds <value>                          # Duration of last scrape in seconds
pskz_scrape_success <value>                                   # Whether last scrape was successful (1 = success)
pskz_last_scrape_error{error_type="balance_fetch_error"} <value>  # Error in balance fetch (1 = error)
pskz_last_scrape_error{error_type="domains_fetch_error"} <value>  # Error in domains fetch (1 = error)
pskz_last_scrape_error{error_type="vps_servers_fetch_error"} <value>  # Error in VPS servers fetch (1 = error)
pskz_last_scrape_error{error_type="k8s_clusters_fetch_error"} <value>  # Error in K8S clusters fetch (1 = error)
pskz_last_scrape_error{error_type="k8s_projects_fetch_error"} <value>  # Error in K8S projects fetch (1 = error)
pskz_last_scrape_error{error_type="lbaas_loadbalancers_fetch_error"} <value>  # Error in LBaaS fetch (1 = error)
pskz_last_scrape_error{error_type="cloud_resources_fetch_error"} <value>  # Error in cloud resources fetch (1 = error)
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Building Docker Image

```bash
make docker-build
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- PS.KZ for providing the GraphQL API
- Prometheus team for the excellent monitoring system
- The Go community for the amazing tools and libraries

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes.