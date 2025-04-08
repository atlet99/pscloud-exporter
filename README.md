# PSCloud Exporter

A Prometheus exporter for PS.KZ (PSCloud) services that collects metrics about your account balance, domains, nameservers, and invoices.

## Features

- Account balance metrics (prepay, credit, debt)
- Domain metrics (expiry dates, status)
- Nameserver metrics (status, IP count)
- Domain price metrics
- Invoice metrics (total amounts, status)
- Comprehensive error reporting
- OAuth 2.0 authentication with PKCE
- Support for both .yml and .yaml configuration files
- Docker support

## Prerequisites

- Go 1.21 or higher
- PS.KZ API credentials (different from your personal account credentials)
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

Create a configuration file (`config.yml` or `config.yaml`) with your PS.KZ API credentials:

```yaml
pscloud:
  username: "your-api-username"
  password: "your-api-password"
  base_url: "https://api.ps.kz/v1"  # Optional, defaults to https://api.ps.kz/v1
  use_http: false                   # Optional, defaults to false
```

## Usage

### Running Locally

```bash
./bin/pscloud-exporter [flags]
```

Available flags:
- `-config.file`: Path to configuration file (default: "config.yml")
- `-web.listen-address`: Address to listen on for web interface and telemetry (default: ":9116")
- `-web.metrics-path`: Path under which to expose metrics (default: "/metrics")

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

### Domain Metrics
- `pskz_domain_expiry_days`: Days until domain expiry
- `pskz_domain_status`: Domain status (1 = active, 0 = inactive)
- `pskz_domain_price`: Domain prices for different operations and zones

### Nameserver Metrics
- `pskz_ns_status`: Nameserver status (1 = active, 0 = inactive)
- `pskz_ns_ip_count`: Number of IPs associated with nameserver

### Invoice Metrics
- `pskz_invoice_total`: Total amount of the invoice
- `pskz_invoice_status`: Invoice status (1 = paid, 0 = unpaid)
- `pskz_invoice_item_amount`: Amount of each item in the invoice

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

- PS.KZ for providing the API
- Prometheus team for the excellent monitoring system
- The Go community for the amazing tools and libraries

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes.