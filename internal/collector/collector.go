package collector

import (
	"log"
	"strconv"
	"time"

	"github.com/atlet99/pscloud-exporter/internal/client"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client *client.Client

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

func New(c *client.Client) *Exporter {
	return &Exporter{
		client: c,

		// Balance metrics
		prepayMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_prepay_balance",
				Help: "Current prepay balance",
			},
			[]string{},
		),
		creditMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_credit_balance",
				Help: "Current credit balance",
			},
			[]string{},
		),
		debtMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_debt_balance",
				Help: "Current debt balance",
			},
			[]string{},
		),

		// Domain metrics
		domainExpiryMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_domain_expiry_days",
				Help: "Days until domain expiry",
			},
			[]string{"domain"},
		),
		domainStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_domain_status",
				Help: "Domain status (1 = active, 0 = inactive)",
			},
			[]string{"domain", "status"},
		),
		domainPriceMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_domain_price",
				Help: "Domain prices",
			},
			[]string{"operation", "zone"},
		),
		nsStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_ns_status",
				Help: "Nameserver status (1 = active, 0 = inactive)",
			},
			[]string{"host", "status"},
		),
		nsIPCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_ns_ip_count",
				Help: "Number of IPs associated with nameserver",
			},
			[]string{"host"},
		),

		// Invoice metrics
		invoiceTotalMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_invoice_total",
				Help: "Total amount of the invoice",
			},
			[]string{"invoice_id", "status", "payment_method"},
		),
		invoiceStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_invoice_status",
				Help: "Invoice status (1 = paid, 0 = unpaid)",
			},
			[]string{"invoice_id", "status"},
		),
		invoiceItemAmountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_invoice_item_amount",
				Help: "Amount of each item in the invoice",
			},
			[]string{"invoice_id", "description"},
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
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

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// Collect balance metrics
	if balance, err := e.client.GetBalance(); err == nil {
		prepay := parseToFloat(balance.Answer.Prepay)
		credit := parseToFloat(balance.Answer.Credit)
		debt := parseToFloat(balance.Answer.CreditInfo.Debt)

		e.prepayMetric.WithLabelValues().Set(prepay)
		e.creditMetric.WithLabelValues().Set(credit)
		e.debtMetric.WithLabelValues().Set(debt)

		e.prepayMetric.Collect(ch)
		e.creditMetric.Collect(ch)
		e.debtMetric.Collect(ch)
	}

	// Try to get domain list using client API first
	domains, err := e.client.GetClientDomainList()
	if err != nil {
		log.Printf("Error getting client domain list: %v, trying domain API", err)
		// If client API fails, try domain API
		domains, err = e.client.GetDomainList()
		if err != nil {
			log.Printf("Error getting domain list: %v", err)
			return
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

			e.nsStatusMetric.WithLabelValues(nsInfo.Answer.Host, nsInfo.Answer.Status).Set(nsStatus)
			e.nsIPCountMetric.WithLabelValues(nsInfo.Answer.Host).Set(float64(len(nsInfo.Answer.IPs)))
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

	e.domainExpiryMetric.Collect(ch)
	e.domainStatusMetric.Collect(ch)
	e.domainPriceMetric.Collect(ch)
	e.nsStatusMetric.Collect(ch)
	e.nsIPCountMetric.Collect(ch)
	e.invoiceTotalMetric.Collect(ch)
	e.invoiceStatusMetric.Collect(ch)
	e.invoiceItemAmountMetric.Collect(ch)
}

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
