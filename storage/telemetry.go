package storage

import (
	"github.com/mlposey/z4/telemetry"
	"go.uber.org/zap"
	"time"
)

type reporter struct {
	close  chan bool
	client *BadgerClient
}

func newReporter(client *BadgerClient) *reporter {
	return &reporter{
		client: client,
		close:  make(chan bool),
	}
}

func (r *reporter) StartReporter() {
	tick := time.Tick(time.Second * 10)
	for {
		select {
		case <-r.close:
			return

		case <-tick:
			r.report()
		}
	}
}

func (r *reporter) report() {
	r.reportNamespaceSizes()
	// TODO: Add more reporting
}

func (r *reporter) reportNamespaceSizes() {
	ns := NewNamespaceStore(r.client)
	namespaces, err := ns.GetAll()
	if err != nil {
		telemetry.Logger.Error("could not get namespaces for reporting",
			zap.Error(err))
		return
	}

	for _, ns := range namespaces {
		fifoPrefix := getFifoPrefix(ns.GetId())
		fifoSize, _ := r.client.DB.EstimateSize(fifoPrefix)

		schedPrefix := getSchedPrefix(ns.GetId())
		schedSize, _ := r.client.DB.EstimateSize(schedPrefix)

		telemetry.NamespaceSize.
			WithLabelValues(ns.GetId()).
			Set(float64(fifoSize + schedSize))
	}
}

func (r *reporter) Close() error {
	r.close <- true
	return nil
}
