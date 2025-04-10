# PSCloud Exporter

A Prometheus exporter for PS.KZ (PSCloud) services that collects metrics about your account balance, domains, and server resources via GraphQL API.

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

## Configuration

Create a configuration file (`config.yml` or `config.yaml`) with your PS.KZ API token:

```yaml
# PSCloud Exporter Configuration
token: "your_token"  # Access token for PS.KZ application
serviceId: "12345"   # Service ID for VPC and VPS API requests (optional)
baseUrl: "https://console.ps.kz"  # Base URL for PS.KZ API (optional)

# Web server configuration
web:
  listenAddress: ":9116"
  metricsPrefix: "pskz"
  telemetryPath: "/metrics"
```

### Obtaining a Token

1. Log in to your PS.KZ account at https://console.ps.kz
2. Navigate to Account Settings
3. Go to "API Integration" or "API Tokens" section
4. Create a new token with appropriate permissions
5. Copy the token and paste it into your configuration file

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

### Account Metrics
- `pskz_prepay_balance`: Current prepay balance
- `pskz_credit_balance`: Current credit balance
- `pskz_debt_balance`: Current debt balance
- `pskz_bonus_balance`: Current bonus balance
- `pskz_blocked_balance`: Current blocked balance

### Domain Metrics
- `pskz_domain_expiry_days`: Days until domain expiry
- `pskz_domain_status`: Domain status (1 = active, 0 = inactive)
- `pskz_domain_counters`: Domain counters (total, active, expired, pending)

### Project Metrics
- `pskz_project_amount`: Project amount
- `pskz_project_disk_usage_gb`: Project disk usage in GB
- `pskz_project_disk_limit_gb`: Project disk limit in GB
- `pskz_project_bw_usage_gb`: Project bandwidth usage in GB
- `pskz_project_bw_limit_gb`: Project bandwidth limit in GB

### Server Metrics
- `pskz_server_ram_mb`: Server RAM in MB
- `pskz_server_cores`: Server CPU cores
- `pskz_server_status`: Server status (1 = active, 0 = inactive)
- `pskz_server_ip_count`: Number of IPs associated with server

### Kubernetes Metrics
- `pskz_k8s_cluster_count`: Number of Kubernetes clusters by status
- `pskz_k8s_cluster_status`: Kubernetes cluster status (1 = active, 0 = inactive)
- `pskz_k8s_cluster_nodes`: Number of worker nodes in Kubernetes cluster
- `pskz_k8s_cluster_masters`: Number of master nodes in Kubernetes cluster
- `pskz_k8s_nodegroup_status`: Kubernetes node group status (1 = active, 0 = inactive)
- `pskz_k8s_nodegroup_nodes`: Number of nodes in Kubernetes node group
- `pskz_k8s_nodegroup_cores`: Number of CPU cores per node in Kubernetes node group
- `pskz_k8s_nodegroup_ram_mb`: Amount of RAM per node in Kubernetes node group (MB)

### LBaaS Metrics
- `pskz_lbaas_loadbalancer_count`: Count of load balancers by status
- `pskz_lbaas_loadbalancer_status`: Load balancer status (1 = active, 0 = inactive)
- `pskz_lbaas_listeners_count`: Count of listeners per load balancer
- `pskz_lbaas_pools_count`: Count of pools per load balancer
- `pskz_lbaas_members_count`: Count of members per load balancer
- `pskz_lbaas_flavor`: Assigned flavor information for load balancer
- `pskz_lbaas_floating_ip`: Whether the load balancer has a floating IP (1 = yes, 0 = no)

### Invoice Metrics
- `pskz_invoice_counters`: Invoice counters (total, unpaid, paid, cancelled)
- `pskz_invoice_amount`: Invoice amount

### Scrape Metrics
- `pskz_scrape_duration_seconds`: Duration of the last scrape in seconds
- `pskz_scrape_success`: Whether the last scrape was successful (1 for success, 0 for failure)
- `pskz_last_scrape_error`: Error status of last scrape attempt (1 if error occurred, with error type label)

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