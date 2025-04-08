package collector

import (
	"strconv"

	"pscloud-exporter/internal/client"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client       *client.Client
	prepayMetric prometheus.Gauge
}

func New(c *client.Client) *Exporter {
	return &Exporter{
		client: c,
		prepayMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pscloud_prepay_balance_tenge",
			Help: "Current prepay balance in KZT",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.prepayMetric.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	balance, err := e.client.GetBalance()
	if err != nil {
		return
	}

	prepay := parseToFloat(balance.Answer.Prepay)
	e.prepayMetric.Set(prepay)
	e.prepayMetric.Collect(ch)
}

func parseToFloat(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)
	return f
}
