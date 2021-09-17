package prometheusConfig

import (
	schedmetrics "github.com/Fishwaldo/go-taskmanager/metrics"
	schedprom "github.com/armon/go-metrics/prometheus"
	//"github.com/prometheus/client_golang/prometheus"
)

func GetPrometicsGaugeConfig() []schedprom.GaugeDefinition {
	var pgd []schedprom.GaugeDefinition
	for _, v := range schedmetrics.MetricsGauges() {
		pgi := schedprom.GaugeDefinition{Name: v.Name, Help: v.Help}
		pgd = append(pgd, pgi)
	}
	return pgd
}

func GetPrometicsCounterConfig() []schedprom.CounterDefinition {
	var pcd []schedprom.CounterDefinition
	for _, v := range schedmetrics.MetricsCounter() {
		pci := schedprom.CounterDefinition{Name: v.Name, Help: v.Help}
		pcd = append(pcd, pci)
	}
	return pcd
}

func GetPrometicsSummaryConfig() []schedprom.SummaryDefinition {
	var psd []schedprom.SummaryDefinition
	for _, v := range schedmetrics.MetricsSummary() {
		psi := schedprom.SummaryDefinition{Name: v.Name, Help: v.Help}
		psd = append(psd, psi)
	}
	return psd
}
