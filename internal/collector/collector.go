package collector

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/atlet99/pscloud-exporter/internal/client"

	kitlog "github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects PS.KZ metrics
type Exporter struct {
	client    *client.Client
	serviceID string // Service ID for VPC and VPS API requests

	// Scrape metrics
	scrapeDurationMetric  prometheus.Gauge
	scrapeSuccessMetric   prometheus.Gauge
	lastScrapeErrorMetric *prometheus.GaugeVec

	// Balance metrics
	prepayMetric  *prometheus.GaugeVec
	creditMetric  *prometheus.GaugeVec
	debtMetric    *prometheus.GaugeVec
	bonusMetric   *prometheus.GaugeVec
	blockedMetric *prometheus.GaugeVec

	// Domain metrics
	domainExpiryMetric   *prometheus.GaugeVec
	domainStatusMetric   *prometheus.GaugeVec
	domainCountersMetric *prometheus.GaugeVec

	// Project metrics
	projectAmountMetric    *prometheus.GaugeVec
	projectDiskUsageMetric *prometheus.GaugeVec
	projectDiskLimitMetric *prometheus.GaugeVec
	projectBwUsageMetric   *prometheus.GaugeVec
	projectBwLimitMetric   *prometheus.GaugeVec

	// Server metrics
	serverRAMMetric     *prometheus.GaugeVec
	serverCoresMetric   *prometheus.GaugeVec
	serverStatusMetric  *prometheus.GaugeVec
	serverIPCountMetric *prometheus.GaugeVec

	// Invoice metrics
	invoiceCountersMetric *prometheus.GaugeVec
	invoiceAmountMetric   *prometheus.GaugeVec

	// Cloud resources metrics
	cloudQuotaMetric        *prometheus.GaugeVec
	cloudSummaryMetric      *prometheus.GaugeVec
	cloudInstanceInfoMetric *prometheus.GaugeVec

	// VPS metrics
	vpsServerStatusMetric     *prometheus.GaugeVec
	vpsServerRamMetric        *prometheus.GaugeVec
	vpsServerCoresMetric      *prometheus.GaugeVec
	vpsServerDiskMetric       *prometheus.GaugeVec
	vpsServerBackupMetric     *prometheus.GaugeVec
	vpsServerIpsProtectMetric *prometheus.GaugeVec
	vpsServerAmountMetric     *prometheus.GaugeVec

	// K8S metrics
	k8sClusterCountMetric    *prometheus.GaugeVec
	k8sClusterStatusMetric   *prometheus.GaugeVec
	k8sClusterNodesMetric    *prometheus.GaugeVec
	k8sClusterMastersMetric  *prometheus.GaugeVec
	k8sNodeGroupStatusMetric *prometheus.GaugeVec
	k8sNodeGroupNodesMetric  *prometheus.GaugeVec
	k8sNodeGroupCoresMetric  *prometheus.GaugeVec
	k8sNodeGroupRAMMetric    *prometheus.GaugeVec

	// LBaaS metrics
	lbaasLoadBalancerCountMetric  *prometheus.GaugeVec
	lbaasLoadBalancerStatusMetric *prometheus.GaugeVec
	lbaasListenersCountMetric     *prometheus.GaugeVec
	lbaasPoolsCountMetric         *prometheus.GaugeVec
	lbaasMembersCountMetric       *prometheus.GaugeVec
	lbaasFlavorMetric             *prometheus.GaugeVec
	lbaasFloatingIPMetric         *prometheus.GaugeVec

	mutex  *sync.Mutex
	logger kitlog.Logger
}

// New creates a new Exporter instance
func New(c *client.Client, serviceID string) *Exporter {
	return &Exporter{
		client:    c,
		serviceID: serviceID,

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
			[]string{"account"},
		),
		creditMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "credit_balance",
				Help:      "Current credit balance",
			},
			[]string{"account"},
		),
		debtMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "debt_balance",
				Help:      "Current debt balance",
			},
			[]string{"account"},
		),
		bonusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "bonus_balance",
				Help:      "Current bonus balance",
			},
			[]string{"account"},
		),
		blockedMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "blocked_balance",
				Help:      "Current blocked balance",
			},
			[]string{"account"},
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
		domainCountersMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "domain_counters",
				Help:      "Domain counters",
			},
			[]string{"domain"},
		),

		// Project metrics
		projectAmountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "project_amount",
				Help:      "Project amount",
			},
			[]string{"project"},
		),
		projectDiskUsageMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "project_disk_usage_gb",
				Help:      "Project disk usage in GB",
			},
			[]string{"project"},
		),
		projectDiskLimitMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "project_disk_limit_gb",
				Help:      "Project disk limit in GB",
			},
			[]string{"project"},
		),
		projectBwUsageMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "project_bw_usage_gb",
				Help:      "Project bandwidth usage in GB",
			},
			[]string{"project"},
		),
		projectBwLimitMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "project_bw_limit_gb",
				Help:      "Project bandwidth limit in GB",
			},
			[]string{"project"},
		),

		// Server metrics
		serverRAMMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "server_ram_mb",
				Help:      "Server RAM in MB",
			},
			[]string{"service_type", "instance_name"},
		),
		serverCoresMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "server_cores",
				Help:      "Server CPU cores",
			},
			[]string{"service_type", "instance_name"},
		),
		serverStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "server_status",
				Help:      "Server status (1 = active, 0 = inactive)",
			},
			[]string{"service_type", "instance_name", "status"},
		),
		serverIPCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "server_ip_count",
				Help:      "Number of IPs associated with server",
			},
			[]string{"service_type", "instance_name"},
		),

		// Invoice metrics
		invoiceCountersMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "invoice_counters",
				Help:      "Invoice counters",
			},
			[]string{"invoice"},
		),
		invoiceAmountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "invoice_amount",
				Help:      "Invoice amount",
			},
			[]string{"invoice"},
		),

		// Cloud resources metrics
		cloudQuotaMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "cloud_quota",
				Help:      "Cloud quota",
			},
			[]string{"resource"},
		),
		cloudSummaryMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "cloud_summary",
				Help:      "Cloud summary",
			},
			[]string{"resource"},
		),
		cloudInstanceInfoMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "cloud_instance_info",
				Help:      "Cloud instance info",
			},
			[]string{"resource", "info"},
		),

		// VPS metrics
		vpsServerStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_status",
				Help:      "VPS server status (1 = active, 0 = inactive)",
			},
			[]string{"instance_name", "status"},
		),
		vpsServerRamMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_ram_mb",
				Help:      "VPS server RAM in MB",
			},
			[]string{"instance_name"},
		),
		vpsServerCoresMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_cores",
				Help:      "VPS server CPU cores",
			},
			[]string{"instance_name"},
		),
		vpsServerDiskMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_disk_gb",
				Help:      "VPS server disk usage in GB",
			},
			[]string{"instance_name"},
		),
		vpsServerBackupMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_backup_gb",
				Help:      "VPS server backup usage in GB",
			},
			[]string{"instance_name"},
		),
		vpsServerIpsProtectMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_ips_protect",
				Help:      "VPS server IPs protect",
			},
			[]string{"instance_name"},
		),
		vpsServerAmountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "vps_server_amount",
				Help:      "VPS server amount",
			},
			[]string{"instance_name"},
		),

		// K8S metrics
		k8sClusterCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_cluster_count",
				Help: "Number of Kubernetes clusters",
			},
			[]string{"status"},
		),
		k8sClusterStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_cluster_status",
				Help: "Status of Kubernetes cluster (1=active, 0=inactive)",
			},
			[]string{"cluster_id", "name", "status", "endpoint_id", "region_id", "project_id", "template_name"},
		),
		k8sClusterNodesMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_cluster_nodes",
				Help: "Number of worker nodes in Kubernetes cluster",
			},
			[]string{"cluster_id", "name"},
		),
		k8sClusterMastersMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_cluster_masters",
				Help: "Number of master nodes in Kubernetes cluster",
			},
			[]string{"cluster_id", "name"},
		),
		k8sNodeGroupStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_nodegroup_status",
				Help: "Status of Kubernetes node group (1=active, 0=inactive)",
			},
			[]string{"cluster_id", "cluster_name", "nodegroup_id", "nodegroup_name", "status"},
		),
		k8sNodeGroupNodesMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_nodegroup_nodes",
				Help: "Number of nodes in Kubernetes node group",
			},
			[]string{"cluster_id", "cluster_name", "nodegroup_id", "nodegroup_name"},
		),
		k8sNodeGroupCoresMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_nodegroup_cores",
				Help: "Number of CPU cores per node in Kubernetes node group",
			},
			[]string{"cluster_id", "cluster_name", "nodegroup_id", "nodegroup_name"},
		),
		k8sNodeGroupRAMMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pskz_k8s_nodegroup_ram_mb",
				Help: "Amount of RAM per node in Kubernetes node group (MB)",
			},
			[]string{"cluster_id", "cluster_name", "nodegroup_id", "nodegroup_name"},
		),

		// LBaaS metrics
		lbaasLoadBalancerCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_loadbalancer_count",
				Help:      "Count of LBaaS load balancers by status",
			},
			[]string{"status"},
		),
		lbaasLoadBalancerStatusMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_loadbalancer_status",
				Help:      "Status of LBaaS load balancer (1 = active, 0 = inactive)",
			},
			[]string{"id", "name", "region_id", "cluster", "status", "vip_address", "floating_ip"},
		),
		lbaasListenersCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_listeners_count",
				Help:      "Count of LBaaS listeners per load balancer",
			},
			[]string{"loadbalancer_id", "loadbalancer_name"},
		),
		lbaasPoolsCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_pools_count",
				Help:      "Count of LBaaS pools per load balancer",
			},
			[]string{"loadbalancer_id", "loadbalancer_name"},
		),
		lbaasMembersCountMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_members_count",
				Help:      "Count of LBaaS members per load balancer",
			},
			[]string{"loadbalancer_id", "loadbalancer_name"},
		),
		lbaasFlavorMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_flavor",
				Help:      "LBaaS flavor information",
			},
			[]string{"loadbalancer_id", "loadbalancer_name", "flavor"},
		),
		lbaasFloatingIPMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "pskz",
				Name:      "lbaas_floating_ip",
				Help:      "Whether the LBaaS has a floating IP (1 = yes, 0 = no)",
			},
			[]string{"loadbalancer_id", "loadbalancer_name"},
		),

		mutex:  &sync.Mutex{},
		logger: kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(log.Writer())),
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
	e.bonusMetric.Describe(ch)
	e.blockedMetric.Describe(ch)
	e.domainExpiryMetric.Describe(ch)
	e.domainStatusMetric.Describe(ch)
	e.domainCountersMetric.Describe(ch)
	e.projectAmountMetric.Describe(ch)
	e.projectDiskUsageMetric.Describe(ch)
	e.projectDiskLimitMetric.Describe(ch)
	e.projectBwUsageMetric.Describe(ch)
	e.projectBwLimitMetric.Describe(ch)
	e.serverRAMMetric.Describe(ch)
	e.serverCoresMetric.Describe(ch)
	e.serverStatusMetric.Describe(ch)
	e.serverIPCountMetric.Describe(ch)
	e.invoiceCountersMetric.Describe(ch)
	e.invoiceAmountMetric.Describe(ch)
	e.cloudQuotaMetric.Describe(ch)
	e.cloudSummaryMetric.Describe(ch)
	e.cloudInstanceInfoMetric.Describe(ch)
	e.vpsServerStatusMetric.Describe(ch)
	e.vpsServerRamMetric.Describe(ch)
	e.vpsServerCoresMetric.Describe(ch)
	e.vpsServerDiskMetric.Describe(ch)
	e.vpsServerBackupMetric.Describe(ch)
	e.vpsServerIpsProtectMetric.Describe(ch)
	e.vpsServerAmountMetric.Describe(ch)
	e.k8sClusterCountMetric.Describe(ch)
	e.k8sClusterStatusMetric.Describe(ch)
	e.k8sClusterNodesMetric.Describe(ch)
	e.k8sClusterMastersMetric.Describe(ch)
	e.k8sNodeGroupStatusMetric.Describe(ch)
	e.k8sNodeGroupNodesMetric.Describe(ch)
	e.k8sNodeGroupCoresMetric.Describe(ch)
	e.k8sNodeGroupRAMMetric.Describe(ch)
	e.lbaasLoadBalancerCountMetric.Describe(ch)
	e.lbaasLoadBalancerStatusMetric.Describe(ch)
	e.lbaasListenersCountMetric.Describe(ch)
	e.lbaasPoolsCountMetric.Describe(ch)
	e.lbaasMembersCountMetric.Describe(ch)
	e.lbaasFlavorMetric.Describe(ch)
	e.lbaasFloatingIPMetric.Describe(ch)
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		e.scrapeDurationMetric.Set(duration)
	}()

	// Reset all metrics before collecting new data
	e.prepayMetric.Reset()
	e.creditMetric.Reset()
	e.debtMetric.Reset()
	e.bonusMetric.Reset()
	e.blockedMetric.Reset()
	e.domainExpiryMetric.Reset()
	e.domainStatusMetric.Reset()
	e.domainCountersMetric.Reset()
	e.projectAmountMetric.Reset()
	e.projectDiskUsageMetric.Reset()
	e.projectDiskLimitMetric.Reset()
	e.projectBwUsageMetric.Reset()
	e.projectBwLimitMetric.Reset()
	e.serverRAMMetric.Reset()
	e.serverCoresMetric.Reset()
	e.serverStatusMetric.Reset()
	e.serverIPCountMetric.Reset()
	e.invoiceCountersMetric.Reset()
	e.invoiceAmountMetric.Reset()
	e.cloudQuotaMetric.Reset()
	e.cloudSummaryMetric.Reset()
	e.cloudInstanceInfoMetric.Reset()
	e.vpsServerStatusMetric.Reset()
	e.vpsServerRamMetric.Reset()
	e.vpsServerCoresMetric.Reset()
	e.vpsServerDiskMetric.Reset()
	e.vpsServerBackupMetric.Reset()
	e.vpsServerIpsProtectMetric.Reset()
	e.vpsServerAmountMetric.Reset()
	e.k8sClusterCountMetric.Reset()
	e.k8sClusterStatusMetric.Reset()
	e.k8sClusterNodesMetric.Reset()
	e.k8sClusterMastersMetric.Reset()
	e.k8sNodeGroupStatusMetric.Reset()
	e.k8sNodeGroupNodesMetric.Reset()
	e.k8sNodeGroupCoresMetric.Reset()
	e.k8sNodeGroupRAMMetric.Reset()
	e.lbaasLoadBalancerCountMetric.Reset()
	e.lbaasLoadBalancerStatusMetric.Reset()
	e.lbaasListenersCountMetric.Reset()
	e.lbaasPoolsCountMetric.Reset()
	e.lbaasMembersCountMetric.Reset()
	e.lbaasFlavorMetric.Reset()
	e.lbaasFloatingIPMetric.Reset()

	// Collect information about balance
	balanceData, err := e.client.GetAccountBalance()
	if err != nil {
		log.Printf("Error getting extended account balance: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("extended_balance_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("extended_balance_fetch_error").Set(0)
		e.processAccountBalanceInfo(balanceData)
	}

	// Alternative method for getting the balance (in case the previous one didn't work)
	balance, err := e.client.GetBalance()
	if err != nil {
		log.Printf("Error getting balance: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("balance_fetch_error").Set(1)
		e.scrapeSuccessMetric.Set(0)

		// Collect error metrics
		e.scrapeDurationMetric.Collect(ch)
		e.scrapeSuccessMetric.Collect(ch)
		e.lastScrapeErrorMetric.Collect(ch)
		return
	}
	e.lastScrapeErrorMetric.WithLabelValues("balance_fetch_error").Set(0)

	e.prepayMetric.WithLabelValues("default").Set(balance.Data.Account.Balance.Prepay)
	e.creditMetric.WithLabelValues("default").Set(balance.Data.Account.Balance.Credit)
	e.debtMetric.WithLabelValues("default").Set(balance.Data.Account.Balance.Debt)

	// Collect domain counters
	domainCounters, err := e.client.GetDomainCounters()
	if err != nil {
		log.Printf("Error getting domain counters: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("domain_counters_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("domain_counters_fetch_error").Set(0)
		e.processDomainCounters(domainCounters)
	}

	// Collect information about domains
	domains, err := e.client.GetDomains()
	if err != nil {
		log.Printf("Error getting domains: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("domains_fetch_error").Set(1)
		e.scrapeSuccessMetric.Set(0)

		// Collect error metrics
		e.scrapeDurationMetric.Collect(ch)
		e.scrapeSuccessMetric.Collect(ch)
		e.lastScrapeErrorMetric.Collect(ch)
		e.prepayMetric.Collect(ch)
		e.creditMetric.Collect(ch)
		e.debtMetric.Collect(ch)
		return
	}
	e.lastScrapeErrorMetric.WithLabelValues("domains_fetch_error").Set(0)

	for _, domain := range domains.Data.Domains.Items {
		expiryTime, err := time.Parse("2006-01-02", domain.ExpiryDate)
		if err != nil {
			log.Printf("Error parsing expiry date for domain %s: %v", domain.Name, err)
			continue
		}

		// Calculate the number of days until expiration
		daysUntilExpiry := time.Until(expiryTime).Hours() / 24
		e.domainExpiryMetric.WithLabelValues(domain.Name).Set(daysUntilExpiry)

		var status float64
		switch domain.Status {
		case "active":
			status = 1
		case "expired":
			status = 0
		default:
			status = -1
		}
		e.domainStatusMetric.WithLabelValues(domain.Name, domain.Status).Set(status)
	}

	// Collect information about projects
	projectsData, err := e.client.GetProjects([]string{"Active"}, 100)
	if err != nil {
		log.Printf("Error getting projects: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("projects_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("projects_fetch_error").Set(0)
		e.processProjectsInfo(projectsData)
	}

	// Collect information about invoices
	invoicesData, err := e.client.GetInvoices("Unpaid", 20)
	if err != nil {
		log.Printf("Error getting invoices: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("invoices_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("invoices_fetch_error").Set(0)
		e.processInvoicesInfo(invoicesData)
	}

	// Collect information about cloud resources
	cloudResources, err := e.client.GetCloudResources()
	if err != nil {
		log.Printf("Error getting cloud resources: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("cloud_resources_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("cloud_resources_fetch_error").Set(0)
		e.processCloudResources(cloudResources)
	}

	// Collect detailed information about cloud instances
	cloudInstances, err := e.client.GetCloudInstances()
	if err != nil {
		log.Printf("Error getting cloud instances: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("cloud_instances_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("cloud_instances_fetch_error").Set(0)
		e.processCloudInstances(cloudInstances)
	}

	// Collect information about VPS servers
	vpsData, err := e.client.GetVpsServersStatus()
	if err != nil {
		log.Printf("Error getting VPS server status: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("vps_servers_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("vps_servers_fetch_error").Set(0)
		e.processVpsServersStatus(vpsData)
	}

	// If service ID is specified, collect information about VPC servers
	if e.serviceID != "" {
		// Collect information about VPC servers
		vpcServers, err := e.client.GetCloudServers(e.serviceID)
		if err != nil {
			log.Printf("Error getting VPC servers: %v", err)
			e.lastScrapeErrorMetric.WithLabelValues("vpc_servers_fetch_error").Set(1)
		} else {
			e.lastScrapeErrorMetric.WithLabelValues("vpc_servers_fetch_error").Set(0)
			e.processServerInfo(vpcServers, "vpc")
		}

		// Collect information about VPS servers
		vpsServers, err := e.client.GetVPSServers(e.serviceID)
		if err != nil {
			log.Printf("Error getting VPS servers: %v", err)
			e.lastScrapeErrorMetric.WithLabelValues("vps_servers_fetch_error").Set(1)
		} else {
			e.lastScrapeErrorMetric.WithLabelValues("vps_servers_fetch_error").Set(0)
			e.processServerInfo(vpsServers, "vps")
		}
	}

	// Collect information about Kubernetes clusters
	k8sClusters, err := e.client.GetK8SClusters()
	if err != nil {
		log.Printf("Error getting K8S clusters: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("k8s_clusters_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("k8s_clusters_fetch_error").Set(0)
		e.processK8SClusters(k8sClusters)
	}

	// Collect information about Kubernetes projects
	k8sProjects, err := e.client.GetK8SProjects()
	if err != nil {
		log.Printf("Error getting K8S projects: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("k8s_projects_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("k8s_projects_fetch_error").Set(0)
		e.processK8SProjects(k8sProjects, ch)
	}

	// Collect information about LBaaS load balancers
	lbaasData, err := e.client.GetLBaaSLoadBalancers()
	if err != nil {
		log.Printf("Error getting LBaaS load balancers: %v", err)
		e.lastScrapeErrorMetric.WithLabelValues("lbaas_loadbalancers_fetch_error").Set(1)
	} else {
		e.lastScrapeErrorMetric.WithLabelValues("lbaas_loadbalancers_fetch_error").Set(0)
		e.processLBaaSData(lbaasData)
	}

	e.scrapeSuccessMetric.Set(1)

	// Collect all metrics
	e.scrapeDurationMetric.Collect(ch)
	e.scrapeSuccessMetric.Collect(ch)
	e.lastScrapeErrorMetric.Collect(ch)
	e.prepayMetric.Collect(ch)
	e.creditMetric.Collect(ch)
	e.debtMetric.Collect(ch)
	e.bonusMetric.Collect(ch)
	e.blockedMetric.Collect(ch)
	e.domainExpiryMetric.Collect(ch)
	e.domainStatusMetric.Collect(ch)
	e.domainCountersMetric.Collect(ch)
	e.projectAmountMetric.Collect(ch)
	e.projectDiskUsageMetric.Collect(ch)
	e.projectDiskLimitMetric.Collect(ch)
	e.projectBwUsageMetric.Collect(ch)
	e.projectBwLimitMetric.Collect(ch)
	e.serverRAMMetric.Collect(ch)
	e.serverCoresMetric.Collect(ch)
	e.serverStatusMetric.Collect(ch)
	e.serverIPCountMetric.Collect(ch)
	e.invoiceCountersMetric.Collect(ch)
	e.invoiceAmountMetric.Collect(ch)
	e.cloudQuotaMetric.Collect(ch)
	e.cloudSummaryMetric.Collect(ch)
	e.cloudInstanceInfoMetric.Collect(ch)
	e.vpsServerStatusMetric.Collect(ch)
	e.vpsServerRamMetric.Collect(ch)
	e.vpsServerCoresMetric.Collect(ch)
	e.vpsServerDiskMetric.Collect(ch)
	e.vpsServerBackupMetric.Collect(ch)
	e.vpsServerIpsProtectMetric.Collect(ch)
	e.vpsServerAmountMetric.Collect(ch)
	e.k8sClusterCountMetric.Collect(ch)
	e.k8sClusterStatusMetric.Collect(ch)
	e.k8sClusterNodesMetric.Collect(ch)
	e.k8sClusterMastersMetric.Collect(ch)
	e.k8sNodeGroupStatusMetric.Collect(ch)
	e.k8sNodeGroupNodesMetric.Collect(ch)
	e.k8sNodeGroupCoresMetric.Collect(ch)
	e.k8sNodeGroupRAMMetric.Collect(ch)
	e.lbaasLoadBalancerCountMetric.Collect(ch)
	e.lbaasLoadBalancerStatusMetric.Collect(ch)
	e.lbaasListenersCountMetric.Collect(ch)
	e.lbaasPoolsCountMetric.Collect(ch)
	e.lbaasMembersCountMetric.Collect(ch)
	e.lbaasFlavorMetric.Collect(ch)
	e.lbaasFloatingIPMetric.Collect(ch)
}

// processAccountBalanceInfo processes account balance information
func (e *Exporter) processAccountBalanceInfo(balanceData map[string]interface{}) {
	// Unpack nested objects
	data, ok := balanceData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for account balance: data field missing")
		return
	}

	account, ok := data["account"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for account balance: account field missing")
		return
	}

	current, ok := account["current"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for account balance: current field missing")
		return
	}

	info, ok := current["info"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for account balance: info field missing")
		return
	}

	// Set balance metrics
	if balance, ok := info["balance"].(float64); ok {
		e.prepayMetric.WithLabelValues("account").Set(balance)
	}

	if bonuses, ok := info["bonuses"].(float64); ok {
		e.bonusMetric.WithLabelValues("account").Set(bonuses)
	}

	if blocked, ok := info["blocked"].(float64); ok {
		e.blockedMetric.WithLabelValues("account").Set(blocked)
	}

	// Process credit
	if credit, ok := info["credit"].(map[string]interface{}); ok {
		if creditVal, ok := credit["credit"].(float64); ok {
			e.creditMetric.WithLabelValues("account_credit").Set(creditVal)
		}

		if maxCredit, ok := credit["maxCredit"].(float64); ok {
			e.creditMetric.WithLabelValues("account_max_credit").Set(maxCredit)
		}

		if availableCredit, ok := credit["availableCredit"].(float64); ok {
			e.creditMetric.WithLabelValues("account_available_credit").Set(float64(availableCredit))
		}
	}
}

// processDomainCounters processes domain counters
func (e *Exporter) processDomainCounters(domainCountersData map[string]interface{}) {
	// Unpack nested objects
	data, ok := domainCountersData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for domain counters: data field missing")
		return
	}

	account, ok := data["account"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for domain counters: account field missing")
		return
	}

	domains, ok := account["domains"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for domain counters: domains field missing")
		return
	}

	stats, ok := domains["stats"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for domain counters: stats field missing")
		return
	}

	// Set domain counter metrics
	if total, ok := stats["total"].(float64); ok {
		e.domainCountersMetric.WithLabelValues("total").Set(total)
	}

	if active, ok := stats["active"].(float64); ok {
		e.domainCountersMetric.WithLabelValues("active").Set(active)
	}

	if expired, ok := stats["expired"].(float64); ok {
		e.domainCountersMetric.WithLabelValues("expired").Set(expired)
	}

	if pending, ok := stats["pending"].(float64); ok {
		e.domainCountersMetric.WithLabelValues("pending").Set(pending)
	}
}

// processProjectsInfo processes information about projects
func (e *Exporter) processProjectsInfo(projectsData map[string]interface{}) {
	// Unpack nested objects
	data, ok := projectsData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for projects: data field missing")
		return
	}

	account, ok := data["account"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for projects: account field missing")
		return
	}

	services, ok := account["services"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for projects: services field missing")
		return
	}

	pagination, ok := services["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for projects: pagination field missing")
		return
	}

	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for projects: items field missing or not an array")
		return
	}

	// Set project metrics
	for _, item := range items {
		projectItem, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid project item: not an object")
			continue
		}

		// Get project ID
		projectId, ok := projectItem["id"].(float64)
		if !ok {
			log.Printf("Invalid project item: id missing or not a number")
			continue
		}

		projectIdStr := fmt.Sprintf("%d", int(projectId))

		// Get project domain for better identification
		domain, ok := projectItem["domain"].(string)
		if ok {
			projectIdStr = fmt.Sprintf("%s-%d", domain, int(projectId))
		}

		// Set project metrics
		if price, ok := projectItem["price"].(float64); ok {
			e.projectAmountMetric.WithLabelValues(projectIdStr).Set(price)
		}

		if diskUsage, ok := projectItem["diskUsage"].(float64); ok {
			e.projectDiskUsageMetric.WithLabelValues(projectIdStr).Set(diskUsage)
		}

		if diskLimit, ok := projectItem["diskLimit"].(float64); ok {
			e.projectDiskLimitMetric.WithLabelValues(projectIdStr).Set(diskLimit)
		}

		if bandwidthUsage, ok := projectItem["bandwidthUsage"].(float64); ok {
			e.projectBwUsageMetric.WithLabelValues(projectIdStr).Set(bandwidthUsage)
		}

		if bandwidthLimit, ok := projectItem["bandwidthLimit"].(float64); ok {
			e.projectBwLimitMetric.WithLabelValues(projectIdStr).Set(bandwidthLimit)
		}
	}
}

// processInvoicesInfo processes information about invoices
func (e *Exporter) processInvoicesInfo(invoicesData map[string]interface{}) {
	// Unpack nested objects
	data, ok := invoicesData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for invoices: data field missing")
		return
	}

	account, ok := data["account"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for invoices: account field missing")
		return
	}

	invoice, ok := account["invoice"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for invoices: invoice field missing")
		return
	}

	// Process counters from invoice data
	counters, ok := invoice["counters"].(map[string]interface{})
	if ok {
		if total, ok := counters["total"].(float64); ok {
			e.invoiceCountersMetric.WithLabelValues("total").Set(total)
		}

		if unpaid, ok := counters["unpaid"].(float64); ok {
			e.invoiceCountersMetric.WithLabelValues("unpaid").Set(unpaid)
		}

		if paid, ok := counters["paid"].(float64); ok {
			e.invoiceCountersMetric.WithLabelValues("paid").Set(paid)
		}

		if cancelled, ok := counters["cancelled"].(float64); ok {
			e.invoiceCountersMetric.WithLabelValues("cancelled").Set(cancelled)
		}
	}

	// Process invoices
	pagination, ok := invoice["pagination"].(map[string]interface{})
	if ok {
		items, ok := pagination["items"].([]interface{})
		if ok {
			// Process each invoice
			for _, item := range items {
				invoiceItem, ok := item.(map[string]interface{})
				if !ok {
					log.Printf("Invalid invoice item: not an object")
					continue
				}

				// Get invoice ID
				invoiceId, ok := invoiceItem["id"].(float64)
				if !ok {
					log.Printf("Invalid invoice item: id missing or not a number")
					continue
				}

				invoiceIdStr := fmt.Sprintf("%d", int(invoiceId))

				// Set invoice metrics
				if total, ok := invoiceItem["total"].(float64); ok {
					e.invoiceAmountMetric.WithLabelValues(invoiceIdStr).Set(total)
				}
			}
		}
	}
}

// processServerInfo processes server information from API response
func (e *Exporter) processServerInfo(serverData map[string]interface{}, serviceType string) {
	// Extract information from GraphQL response data
	// Response structure: {"data": {"vpc": {"instance": {"pagination": {"items": [...]}}}}}
	data, ok := serverData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for server info: data field missing")
		return
	}

	vpc, ok := data["vpc"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for server info: vpc field missing")
		return
	}

	instance, ok := vpc["instance"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for server info: instance field missing")
		return
	}

	pagination, ok := instance["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for server info: pagination field missing")
		return
	}

	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for server info: items field missing or not an array")
		return
	}

	for _, item := range items {
		server, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid server item: not an object")
			continue
		}

		instanceName, ok := server["instanceName"].(string)
		if !ok {
			log.Printf("Invalid server item: instanceName missing or not a string")
			continue
		}

		// RAM
		ram, ok := server["ram"].(float64)
		if ok {
			e.serverRAMMetric.WithLabelValues(serviceType, instanceName).Set(ram)
		}

		// Cores
		cores, ok := server["cores"].(float64)
		if ok {
			e.serverCoresMetric.WithLabelValues(serviceType, instanceName).Set(cores)
		}

		// Status
		status, ok := server["status"].(string)
		if ok {
			var statusValue float64
			if status == "ACTIVE" {
				statusValue = 1
			} else {
				statusValue = 0
			}
			e.serverStatusMetric.WithLabelValues(serviceType, instanceName, status).Set(statusValue)
		}

		// IP Addresses
		ips, ok := server["floatingIpsArray"].([]interface{})
		if ok {
			e.serverIPCountMetric.WithLabelValues(serviceType, instanceName).Set(float64(len(ips)))
		}
	}
}

// processCloudResources processes information about cloud resources
func (e *Exporter) processCloudResources(cloudResourcesData map[string]interface{}) {
	// Unpack nested objects
	data, ok := cloudResourcesData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud resources: data field missing")
		return
	}

	vpc, ok := data["vpc"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud resources: vpc field missing")
		return
	}

	service, ok := vpc["service"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud resources: service field missing")
		return
	}

	// Process quotas
	quotas, ok := service["quotas"].(map[string]interface{})
	if ok {
		resources, ok := quotas["resources"].([]interface{})
		if ok {
			for _, res := range resources {
				resource, ok := res.(map[string]interface{})
				if !ok {
					continue
				}

				name, ok := resource["name"].(string)
				if !ok {
					continue
				}

				if used, ok := resource["used"].(float64); ok {
					e.cloudQuotaMetric.WithLabelValues(fmt.Sprintf("%s_used", name)).Set(used)
				}

				if limit, ok := resource["limit"].(float64); ok {
					e.cloudQuotaMetric.WithLabelValues(fmt.Sprintf("%s_limit", name)).Set(limit)
				}
			}
		}
	}

	// Process summary
	summary, ok := service["summary"].(map[string]interface{})
	if ok {
		if cpuCores, ok := summary["cpuCores"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("cpu_cores").Set(cpuCores)
		}

		if ramSizeGb, ok := summary["ramSizeGb"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("ram_gb").Set(ramSizeGb)
		}

		if instancesCount, ok := summary["instancesCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("instances_count").Set(instancesCount)
		}

		if volumesCount, ok := summary["volumesCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("volumes_count").Set(volumesCount)
		}

		if volumesSizeGb, ok := summary["volumesSizeGb"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("volumes_size_gb").Set(volumesSizeGb)
		}

		if networksCount, ok := summary["networksCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("networks_count").Set(networksCount)
		}

		if floatingIpsCount, ok := summary["floatingIpsCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("floating_ips_count").Set(floatingIpsCount)
		}

		if securityGroupsCount, ok := summary["securityGroupsCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("security_groups_count").Set(securityGroupsCount)
		}

		if routersCount, ok := summary["routersCount"].(float64); ok {
			e.cloudSummaryMetric.WithLabelValues("routers_count").Set(routersCount)
		}
	}

	// Process instance information
	instanceInfo, ok := service["instanceInfo"].(map[string]interface{})
	if ok {
		for resource, info := range instanceInfo {
			resourceInfo, ok := info.(map[string]interface{})
			if !ok {
				continue
			}

			for infoKey, infoValue := range resourceInfo {
				e.cloudInstanceInfoMetric.WithLabelValues(resource, infoKey).Set(infoValue.(float64))
			}
		}
	}
}

// processCloudInstances processes detailed information about cloud instances
func (e *Exporter) processCloudInstances(instancesData map[string]interface{}) {
	// Unpack nested objects
	data, ok := instancesData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud instances: data field missing")
		return
	}

	vpc, ok := data["vpc"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud instances: vpc field missing")
		return
	}

	instance, ok := vpc["instance"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud instances: instance field missing")
		return
	}

	pagination, ok := instance["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud instances: pagination field missing")
		return
	}

	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for cloud instances: items field missing or not an array")
		return
	}

	// Process each instance
	for _, item := range items {
		instanceItem, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid instance item: not an object")
			continue
		}

		// Get instance name
		instanceName, ok := instanceItem["instanceName"].(string)
		if !ok {
			log.Printf("Invalid instance item: instanceName missing or not a string")
			continue
		}

		// Set metrics for instance status
		status, ok := instanceItem["status"].(string)
		if ok {
			var statusValue float64
			switch status {
			case "ACTIVE":
				statusValue = 1
			case "SHUTOFF":
				statusValue = 0
			case "PAUSED":
				statusValue = 0.5
			default:
				statusValue = -1
			}
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName, "status").Set(statusValue)
		}

		// Set metrics for flavor
		flavorName, ok := instanceItem["flavorName"].(string)
		if ok {
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName, "flavor_name").Set(1)
			// Save flavor name in label
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName+":"+flavorName, "flavor").Set(1)
		}

		// Count attached volumes
		volumesAttached, ok := instanceItem["volumesAttached"].([]interface{})
		if ok {
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName, "volumes_count").Set(float64(len(volumesAttached)))

			// Count total size of attached volumes
			var totalVolumeSize float64
			for _, vol := range volumesAttached {
				volume, ok := vol.(map[string]interface{})
				if !ok {
					continue
				}

				if volumeSize, ok := volume["volumeSize"].(float64); ok {
					totalVolumeSize += volumeSize
				}
			}
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName, "volumes_total_size").Set(totalVolumeSize)
		}

		// Count IP addresses
		floatingIps, ok := instanceItem["floatingIpsArray"].([]interface{})
		if ok {
			e.cloudInstanceInfoMetric.WithLabelValues(instanceName, "floating_ips_count").Set(float64(len(floatingIps)))
		}
	}
}

// processVpsServersStatus processes information about VPS servers
func (e *Exporter) processVpsServersStatus(vpsData map[string]interface{}) {
	// Unpack nested objects
	data, ok := vpsData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for VPS servers: data field missing")
		return
	}

	vps, ok := data["vps"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for VPS servers: vps field missing")
		return
	}

	server, ok := vps["server"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for VPS servers: server field missing")
		return
	}

	pagination, ok := server["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for VPS servers: pagination field missing")
		return
	}

	// Count servers by status
	statusCounts := make(map[string]int)

	// Process servers
	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for VPS servers: items field missing or not an array")
		return
	}

	for _, item := range items {
		serverInfo, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Get server ID and name
		serverId, _ := serverInfo["serverId"].(float64)
		serverName, _ := serverInfo["name"].(string)
		serverIdStr := fmt.Sprintf("%d", int(serverId))

		// Count status
		status, ok := serverInfo["status"].(string)
		if !ok {
			status = "UNKNOWN"
		}
		statusCounts[status]++

		// Set server status metric (1 if active, 0 otherwise)
		statusValue := 0.0
		if status == "ACTIVE" {
			statusValue = 1.0
		}
		e.vpsServerStatusMetric.WithLabelValues(serverIdStr, serverName, status).Set(statusValue)

		// Get region
		regionId, _ := serverInfo["regionId"].(string)

		// Get tariff info if available
		if tariff, ok := serverInfo["tariff"].(map[string]interface{}); ok {
			// Set RAM metric
			if ram, ok := tariff["ramGb"].(float64); ok {
				e.vpsServerRamMetric.WithLabelValues(serverIdStr, serverName, regionId).Set(ram)
			}

			// Set cores metric
			if cores, ok := tariff["cores"].(float64); ok {
				e.vpsServerCoresMetric.WithLabelValues(serverIdStr, serverName, regionId).Set(cores)
			}
		}
	}

	// Set status counters
	for status, count := range statusCounts {
		e.vpsServerStatusMetric.WithLabelValues("all", "total", status).Set(float64(count))
	}
}

// processK8SClusters processes Kubernetes clusters information
func (e *Exporter) processK8SClusters(k8sClustersData map[string]interface{}) {
	// Unpack nested objects
	data, ok := k8sClustersData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S clusters: data field missing")
		return
	}

	k8saas, ok := data["k8saas"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S clusters: k8saas field missing")
		return
	}

	cluster, ok := k8saas["cluster"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S clusters: cluster field missing")
		return
	}

	pagination, ok := cluster["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S clusters: pagination field missing")
		return
	}

	// Get total count of clusters
	count, ok := pagination["count"].(float64)
	if ok {
		e.k8sClusterCountMetric.WithLabelValues("total").Set(count)
	}

	// Process clusters
	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S clusters: items field missing or not an array")
		return
	}

	// Initialize counters for cluster statuses
	statusCounts := make(map[string]int)

	for _, item := range items {
		clusterItem, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid cluster item: not an object")
			continue
		}

		clusterId, ok := clusterItem["_id"].(string)
		if !ok {
			log.Printf("Invalid cluster item: _id missing or not a string")
			continue
		}

		name, ok := clusterItem["name"].(string)
		if !ok {
			name = "unknown"
		}

		status, ok := clusterItem["status"].(string)
		if !ok {
			status = "unknown"
		}

		// Count clusters by status
		statusCounts[status]++

		projectId := ""
		if pid, ok := clusterItem["projectId"].(float64); ok {
			projectId = fmt.Sprintf("%.0f", pid)
		}

		endpointId, _ := clusterItem["endpointId"].(string)
		regionId, _ := clusterItem["regionId"].(string)

		// Get template name
		templateName := "unknown"
		if template, ok := clusterItem["clusterTemplate"].(map[string]interface{}); ok {
			if tName, ok := template["name"].(string); ok {
				templateName = tName
			}
		}

		// Set cluster status metric (1 for active, 0 for inactive or other)
		var statusValue float64
		if status == "CREATE_COMPLETE" || status == "UPDATE_COMPLETE" {
			statusValue = 1
		} else {
			statusValue = 0
		}

		e.k8sClusterStatusMetric.WithLabelValues(
			clusterId,
			name,
			status,
			endpointId,
			regionId,
			projectId,
			templateName,
		).Set(statusValue)

		// Set node count metrics
		if nodeCount, ok := clusterItem["nodeCount"].(float64); ok {
			e.k8sClusterNodesMetric.WithLabelValues(clusterId, name).Set(nodeCount)
		}

		if masterCount, ok := clusterItem["masterCount"].(float64); ok {
			e.k8sClusterMastersMetric.WithLabelValues(clusterId, name).Set(masterCount)
		}

		// Process node groups
		if nodeGroups, ok := clusterItem["clusterNodeGroups"].([]interface{}); ok {
			for _, ng := range nodeGroups {
				nodeGroup, ok := ng.(map[string]interface{})
				if !ok {
					continue
				}

				nodeGroupId, ok := nodeGroup["_id"].(string)
				if !ok {
					continue
				}

				nodeGroupName, _ := nodeGroup["name"].(string)
				nodeGroupStatus, _ := nodeGroup["status"].(string)

				// Set status metric (1 for active, 0 for inactive or other)
				var nodeGroupStatusValue float64
				if nodeGroupStatus == "CREATE_COMPLETE" || nodeGroupStatus == "UPDATE_COMPLETE" {
					nodeGroupStatusValue = 1
				} else {
					nodeGroupStatusValue = 0
				}

				e.k8sNodeGroupStatusMetric.WithLabelValues(
					clusterId,
					name,
					nodeGroupId,
					nodeGroupName,
					nodeGroupStatus,
				).Set(nodeGroupStatusValue)

				// Set node count for the group
				if nodeCount, ok := nodeGroup["nodeCount"].(float64); ok {
					e.k8sNodeGroupNodesMetric.WithLabelValues(
						clusterId,
						name,
						nodeGroupId,
						nodeGroupName,
					).Set(nodeCount)
				}

				// Process flavor details
				if flavorDetailed, ok := nodeGroup["flavorDetailed"].(map[string]interface{}); ok {
					if vcpus, ok := flavorDetailed["vcpus"].(float64); ok {
						e.k8sNodeGroupCoresMetric.WithLabelValues(
							clusterId,
							name,
							nodeGroupId,
							nodeGroupName,
						).Set(vcpus)
					}

					if ram, ok := flavorDetailed["ram"].(float64); ok {
						e.k8sNodeGroupRAMMetric.WithLabelValues(
							clusterId,
							name,
							nodeGroupId,
							nodeGroupName,
						).Set(ram)
					}
				}
			}
		}
	}

	// Set metrics for cluster counts by status
	for status, count := range statusCounts {
		e.k8sClusterCountMetric.WithLabelValues(status).Set(float64(count))
	}
}

// processLBaaSData processes LBaaS load balancer information
func (e *Exporter) processLBaaSData(lbaasData map[string]interface{}) {
	// Unpack nested objects
	data, ok := lbaasData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for LBaaS data: data field missing")
		return
	}

	lbaas, ok := data["lbaas"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for LBaaS data: lbaas field missing")
		return
	}

	loadBalancer, ok := lbaas["loadBalancer"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for LBaaS data: loadBalancer field missing")
		return
	}

	pagination, ok := loadBalancer["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for LBaaS data: pagination field missing")
		return
	}

	count, ok := pagination["count"].(float64)
	if ok {
		e.lbaasLoadBalancerCountMetric.WithLabelValues("total").Set(count)
	}

	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for LBaaS data: items field missing or not an array")
		return
	}

	// Initialize counters for load balancer statuses
	statusCounts := make(map[string]int)

	for _, item := range items {
		lb, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid load balancer item: not an object")
			continue
		}

		id, ok := lb["_id"].(string)
		if !ok {
			log.Printf("Invalid load balancer item: _id missing or not a string")
			continue
		}

		name, ok := lb["name"].(string)
		if !ok {
			name = "unknown"
		}

		regionID, ok := lb["regionId"].(string)
		if !ok {
			regionID = "unknown"
		}

		vipAddress, ok := lb["vipAddress"].(string)
		if !ok {
			vipAddress = "unknown"
		}

		status, ok := lb["provisioningStatus"].(string)
		if !ok {
			status = "unknown"
		}

		// Count load balancers by status
		statusCounts[status]++

		// Get cluster information
		clusterName := "unknown"
		if cluster, ok := lb["cluster"].(map[string]interface{}); ok {
			if name, ok := cluster["name"].(string); ok {
				clusterName = name
			}
		}

		// Get floating IP
		floatingIP, ok := lb["floatingIpAddress"].(string)
		if !ok {
			floatingIP = ""
		}

		// Set load balancer status metric (1 for active, 0 for inactive or other)
		var statusValue float64
		if status == "ACTIVE" {
			statusValue = 1
		} else {
			statusValue = 0
		}

		e.lbaasLoadBalancerStatusMetric.WithLabelValues(
			id,
			name,
			regionID,
			clusterName,
			status,
			vipAddress,
			floatingIP,
		).Set(statusValue)

		// Set flavor metric
		flavorName, ok := lb["flavorName"].(string)
		if ok && flavorName != "" {
			e.lbaasFlavorMetric.WithLabelValues(id, name, flavorName).Set(1)
		}

		// Set floating IP metric
		if floatingIP != "" {
			e.lbaasFloatingIPMetric.WithLabelValues(id, name).Set(1)
		} else {
			e.lbaasFloatingIPMetric.WithLabelValues(id, name).Set(0)
		}

		// Process listeners
		listeners, ok := lb["listeners"].([]interface{})
		if ok {
			e.lbaasListenersCountMetric.WithLabelValues(id, name).Set(float64(len(listeners)))
		}

		// Process pools
		pools, ok := lb["pools"].([]interface{})
		if ok {
			e.lbaasPoolsCountMetric.WithLabelValues(id, name).Set(float64(len(pools)))
		}

		// Process members
		members, ok := lb["members"].([]interface{})
		if ok {
			e.lbaasMembersCountMetric.WithLabelValues(id, name).Set(float64(len(members)))
		}
	}

	// Set metrics for load balancer counts by status
	for status, count := range statusCounts {
		e.lbaasLoadBalancerCountMetric.WithLabelValues(status).Set(float64(count))
	}
}

// processK8SProjects processes Kubernetes projects information
func (e *Exporter) processK8SProjects(k8sProjectsData map[string]interface{}, ch chan<- prometheus.Metric) {
	// Unpack nested objects
	data, ok := k8sProjectsData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S projects: data field missing")
		return
	}

	k8saas, ok := data["k8saas"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S projects: k8saas field missing")
		return
	}

	project, ok := k8saas["project"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S projects: project field missing")
		return
	}

	pagination, ok := project["pagination"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S projects: pagination field missing")
		return
	}

	// Process items
	items, ok := pagination["items"].([]interface{})
	if !ok {
		log.Printf("Invalid data structure for K8S projects: items field missing or not an array")
		return
	}

	// Initialize counters for project statuses
	statusCounts := make(map[string]int)
	typesCounts := make(map[string]int)

	for _, item := range items {
		projectItem, ok := item.(map[string]interface{})
		if !ok {
			log.Printf("Invalid project item: not an object")
			continue
		}

		// Get project ID and name
		projectId := "unknown"
		if pid, ok := projectItem["projectId"].(string); ok {
			projectId = pid
		} else if pid, ok := projectItem["projectId"].(float64); ok {
			projectId = fmt.Sprintf("%.0f", pid)
		}

		projectName := projectId
		if pname, ok := projectItem["projectName"].(string); ok && pname != "" {
			projectName = pname
		}

		// Get status and type
		status, _ := projectItem["status"].(string)
		projectType, _ := projectItem["type"].(string)

		// Count projects by status and type
		if status != "" {
			statusCounts[status]++
		}

		if projectType != "" {
			typesCounts[projectType]++
		}

		// Process OpenStack services quota
		if openstackServices, ok := projectItem["openstackServices"].([]interface{}); ok {
			for _, service := range openstackServices {
				serviceItem, ok := service.(map[string]interface{})
				if !ok {
					continue
				}

				serviceName, _ := serviceItem["name"].(string)
				regionId, _ := serviceItem["regionId"].(string)

				// Process quota
				if quota, ok := serviceItem["quota"].([]interface{}); ok {
					for _, q := range quota {
						quotaItem, ok := q.(map[string]interface{})
						if !ok {
							continue
						}

						key, ok := quotaItem["key"].(string)
						if !ok {
							continue
						}

						// Set limit metric
						if limit, ok := quotaItem["limit"].(float64); ok {
							name := fmt.Sprintf("pskz_k8s_project_quota_%s_%s_limit", serviceName, key)
							desc := prometheus.NewDesc(
								name,
								fmt.Sprintf("Quota limit for %s %s", serviceName, key),
								[]string{"project_id", "project_name", "region_id"},
								nil,
							)
							ch <- prometheus.MustNewConstMetric(
								desc,
								prometheus.GaugeValue,
								limit,
								projectId, projectName, regionId,
							)
						}

						// Set usage metric
						if inUse, ok := quotaItem["inUse"].(float64); ok {
							name := fmt.Sprintf("pskz_k8s_project_quota_%s_%s_used", serviceName, key)
							desc := prometheus.NewDesc(
								name,
								fmt.Sprintf("Quota usage for %s %s", serviceName, key),
								[]string{"project_id", "project_name", "region_id"},
								nil,
							)
							ch <- prometheus.MustNewConstMetric(
								desc,
								prometheus.GaugeValue,
								inUse,
								projectId, projectName, regionId,
							)
						}
					}
				}
			}
		}
	}

	// Set metrics for project counts by status
	for status, count := range statusCounts {
		name := "pskz_k8s_project_status_count"
		desc := prometheus.NewDesc(
			name,
			"Number of Kubernetes projects by status",
			[]string{"status"},
			nil,
		)
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			status,
		)
	}

	// Set metrics for project counts by type
	for projectType, count := range typesCounts {
		name := "pskz_k8s_project_type_count"
		desc := prometheus.NewDesc(
			name,
			"Number of Kubernetes projects by type",
			[]string{"type"},
			nil,
		)
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			projectType,
		)
	}
}
