package collector

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/atlet99/pscloud-exporter/internal/client"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client *client.Client

	// Scrape metrics
	scrapeDurationMetric  prometheus.Gauge
	scrapeSuccessMetric   prometheus.Gauge
	lastScrapeErrorMetric *prometheus.GaugeVec

	// Balance metrics
	prepayMetric *prometheus.GaugeVec
	creditMetric *prometheus.GaugeVec
	debtMetric   *prometheus.GaugeVec

	// Domain metrics
	domainExpiryMetric *prometheus.GaugeVec
	domainStatusMetric *prometheus.GaugeVec
	domainPriceMetric  *prometheus.GaugeVec
	nsStatusMetric     *prometheus.GaugeVec
	nsIPCountMetric    *prometheus.GaugeVec

	// Invoice metrics
	invoiceTotalMetric      *prometheus.GaugeVec
	invoiceStatusMetric     *prometheus.GaugeVec
	invoiceItemAmountMetric *prometheus.GaugeVec
}

// New creates a new Exporter instance
func New(c *client.Client) *Exporter {
	return &Exporter{
		client: c,

		// Scrape metrics
		scrapeDurationMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "scrape_duration_seconds",
				Help:      "Duration of the last scrape in seconds",
			},
		),
		scrapeSuccessMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "scrape_success",
				Help:      "Whether the last scrape was successful (1 for success, 0 for failure)",
			},
		),
		lastScrapeErrorMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "last_scrape_error",
				Help:      "Error status of last scrape attempt (1 if error occurred, with error type label)",
			},
			[]string{"error_type"},
		),

		// Balance metrics
		prepayMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "prepay_balance",
				Help:      "Current prepay balance",
			},
			[]string{"username"},
		),
		creditMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "credit_balance",
				Help:      "Current credit balance",
			},
			[]string{"username"},
		),
		debtMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "debt_balance",
				Help:      "Current debt balance",
			},
			[]string{"username"},
		),

		// Domain metrics
		domainExpiryMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "domain_expiry_days",
				Help:      "Days until domain expiry",
			},
			[]string{"domain"},
		),
		domainStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "domain_status",
				Help:      "Domain status (1 = active, 0 = inactive)",
			},
			[]string{"domain", "status"},
		),
		domainPriceMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "domain_price",
				Help:      "Domain prices",
			},
			[]string{"operation", "zone"},
		),
		nsStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "ns_status",
				Help:      "Nameserver status (1 = active, 0 = inactive)",
			},
			[]string{"host", "status"},
		),
		nsIPCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "ns_ip_count",
				Help:      "Number of IPs associated with nameserver",
			},
			[]string{"host"},
		),

		// Invoice metrics
		invoiceTotalMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "invoice_total",
				Help:      "Total amount of the invoice",
			},
			[]string{"invoice_id", "status", "payment_method"},
		),
		invoiceStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "invoice_status",
				Help:      "Invoice status (1 = paid, 0 = unpaid)",
			},
			[]string{"invoice_id", "status"},
		),
		invoiceItemAmountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "invoice_item_amount",
				Help:      "Amount of each item in the invoice",
			},
			[]string{"invoice_id", "description"},
		),
	}
}

// Describe implements prometheus.Collector
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.scrapeDurationMetric.Describe(ch)
	e.scrapeSuccessMetric.Describe(ch)
	e.lastScrapeErrorMetric.Describe(ch)
	e.prepayMetric.Describe(ch)
	e.creditMetric.Describe(ch)
	e.debtMetric.Describe(ch)
	e.domainExpiryMetric.Describe(ch)
	e.domainStatusMetric.Describe(ch)
	e.domainPriceMetric.Describe(ch)
	e.nsStatusMetric.Describe(ch)
	e.nsIPCountMetric.Describe(ch)
	e.invoiceTotalMetric.Describe(ch)
	e.invoiceStatusMetric.Describe(ch)
	e.invoiceItemAmountMetric.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		e.scrapeDurationMetric.Set(duration)
		e.scrapeDurationMetric.Collect(ch)
		e.scrapeSuccessMetric.Collect(ch)
		e.lastScrapeErrorMetric.Collect(ch)
	}()

	// Reset all metrics
	e.lastScrapeErrorMetric.Reset()
	e.prepayMetric.Reset()
	e.creditMetric.Reset()
	e.debtMetric.Reset()
	e.domainExpiryMetric.Reset()
	e.domainStatusMetric.Reset()
	e.domainPriceMetric.Reset()
	e.nsStatusMetric.Reset()
	e.nsIPCountMetric.Reset()
	e.invoiceTotalMetric.Reset()
	e.invoiceStatusMetric.Reset()
	e.invoiceItemAmountMetric.Reset()

	balance, err := e.client.GetBalance()
	if err != nil {
		log.Printf("Error getting balance: %v", err)
		e.scrapeSuccessMetric.Set(0)

		if strings.Contains(err.Error(), "authentication failed") {
			e.lastScrapeErrorMetric.WithLabelValues("authentication_error").Set(1)
			// Set balance metrics to -1 on authentication error
			e.prepayMetric.WithLabelValues(e.client.GetUsername()).Set(-1)
			e.creditMetric.WithLabelValues(e.client.GetUsername()).Set(-1)
			e.debtMetric.WithLabelValues(e.client.GetUsername()).Set(-1)
		} else {
			e.lastScrapeErrorMetric.WithLabelValues("balance_fetch_error").Set(1)
		}
		return
	}

	e.scrapeSuccessMetric.Set(1)
	e.prepayMetric.WithLabelValues(e.client.GetUsername()).Set(parseToFloat(balance.Answer.Prepay))
	e.creditMetric.WithLabelValues(e.client.GetUsername()).Set(parseToFloat(balance.Answer.Credit))
	e.debtMetric.WithLabelValues(e.client.GetUsername()).Set(parseToFloat(balance.Answer.CreditInfo.Debt))

	// Try to get domain list using client API first
	domains, err := e.client.GetClientDomainList()
	if err != nil {
		log.Printf("Error getting client domain list: %v, trying domain API", err)
		// If client API fails, try domain API
		domainListResp, err := e.client.GetDomainList()
		if err != nil {
			log.Printf("Error getting domain list: %v", err)
			return
		}
		// Convert DomainListResponse to ClientDomainListResponse
		domains = &client.ClientDomainListResponse{
			Result: domainListResp.Result,
			Answer: domainListResp.Answer,
		}
	}

	// Process domain information
	for _, domain := range domains.Answer {
		// Parse expiry date
		if expiryTime, err := time.Parse("2006-01-02", domain.ExpiryDate); err == nil {
			e.domainExpiryMetric.WithLabelValues(domain.Domain).Set(float64(expiryTime.Unix()))
		}

		// Set domain status
		status := 0.0
		if domain.Status == "Active" {
			status = 1.0
		}
		e.domainStatusMetric.WithLabelValues(domain.Domain, domain.Status).Set(status)

		// Collect NSS information for each domain
		nssInfo, err := e.client.GetDomainNSSList(domain.Domain)
		if err != nil {
			log.Printf("Error getting NSS for domain %s: %v", domain.Domain, err)
			continue
		}

		// For each nameserver, collect its information
		for _, ns := range nssInfo.Answer.Nameservers.NS {
			nsInfo, err := e.client.GetNSInfo(ns)
			if err != nil {
				log.Printf("Error getting NS info for %s: %v", ns, err)
				continue
			}

			nsStatus := 0.0
			if nsInfo.Answer.Status == "active" {
				nsStatus = 1.0
			}
			e.nsStatusMetric.WithLabelValues(ns, nsInfo.Answer.Status).Set(nsStatus)
			e.nsIPCountMetric.WithLabelValues(ns).Set(float64(len(nsInfo.Answer.IPs)))
		}
	}

	// Collect domain prices
	if prices, err := e.client.GetPrices(); err == nil {
		// KZ zone prices
		regPrice := parseToFloat(prices.Answer.ZoneKZ.Reg.Price)
		renewPrice := parseToFloat(prices.Answer.ZoneKZ.Renew.Price)
		e.domainPriceMetric.WithLabelValues("registration", ".kz").Set(regPrice)
		e.domainPriceMetric.WithLabelValues("renewal", ".kz").Set(renewPrice)

		// COM.KZ zone prices
		regPrice = parseToFloat(prices.Answer.ZoneComKZ.Reg.Price)
		renewPrice = parseToFloat(prices.Answer.ZoneComKZ.Renew.Price)
		e.domainPriceMetric.WithLabelValues("registration", ".com.kz").Set(regPrice)
		e.domainPriceMetric.WithLabelValues("renewal", ".com.kz").Set(renewPrice)

		// ORG.KZ zone prices
		regPrice = parseToFloat(prices.Answer.ZoneOrgKZ.Reg.Price)
		renewPrice = parseToFloat(prices.Answer.ZoneOrgKZ.Renew.Price)
		e.domainPriceMetric.WithLabelValues("registration", ".org.kz").Set(regPrice)
		e.domainPriceMetric.WithLabelValues("renewal", ".org.kz").Set(renewPrice)
	}

	// Collect invoice metrics if invoice ID is provided
	if invoiceID := e.getActiveInvoiceID(); invoiceID != "" {
		if invoice, err := e.client.GetInvoiceDetails(invoiceID); err == nil {
			total := parseToFloat(invoice.Answer.Total)
			e.invoiceTotalMetric.WithLabelValues(
				invoice.Answer.ID,
				invoice.Answer.Status,
				invoice.Answer.PaymentMethod,
			).Set(total)

			// Set invoice status
			status := 0.0
			if invoice.Answer.Status == "Paid" {
				status = 1.0
			}
			e.invoiceStatusMetric.WithLabelValues(
				invoice.Answer.ID,
				invoice.Answer.Status,
			).Set(status)

			// Set invoice items
			for _, item := range invoice.Answer.Items {
				amount := parseToFloat(item.Amount)
				e.invoiceItemAmountMetric.WithLabelValues(
					invoice.Answer.ID,
					item.Description,
				).Set(amount)
			}
		}
	}

	// Collect all metrics
	e.prepayMetric.Collect(ch)
	e.creditMetric.Collect(ch)
	e.debtMetric.Collect(ch)
	e.domainExpiryMetric.Collect(ch)
	e.domainStatusMetric.Collect(ch)
	e.domainPriceMetric.Collect(ch)
	e.nsStatusMetric.Collect(ch)
	e.nsIPCountMetric.Collect(ch)
	e.invoiceTotalMetric.Collect(ch)
	e.invoiceStatusMetric.Collect(ch)
	e.invoiceItemAmountMetric.Collect(ch)
}

// parseToFloat converts string value to float64
func parseToFloat(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// getActiveInvoiceID returns the ID of the active invoice that needs to be monitored
// This method should be implemented according to your needs
func (e *Exporter) getActiveInvoiceID() string {
	// TODO: Implement logic to get active invoice ID
	// This could be:
	// 1. Read from configuration
	// 2. Get from environment variable
	// 3. Get from command line flag
	// 4. Get from file
	return ""
}
