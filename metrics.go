package main

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	FailedToAuthenticate prometheus.Counter
	Banned *prometheus.CounterVec
	IpsetEntries prometheus.GaugeFunc
}


func NewMetrics(ipsetEntries func() float64) (*Metrics, error) {
	metrics := Metrics{
		FailedToAuthenticate: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "failed_to_authenticate",
			Help: "Failed to authenticate mentions in log",
		}),
		Banned: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "banned",
			Help: "Banned counter",
		}, []string{`type`}),
		IpsetEntries: prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "ipset_entries",
			Help: "Count of entries in ipset list",
		}, ipsetEntries),
	}
	if err := prometheus.Register(metrics.FailedToAuthenticate); err != nil {
		ErrorLog.Println(err)
		return nil, err
	}
	if err := prometheus.Register(metrics.Banned); err != nil {
		ErrorLog.Println(err)
		return nil, err
	}
	if err := prometheus.Register(metrics.IpsetEntries); err != nil {
		ErrorLog.Println(err)
		return nil, err
	}
	return &metrics, nil
}