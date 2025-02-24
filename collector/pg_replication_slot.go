// Copyright 2023 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const replicationSlotSubsystem = "replication_slot"

func init() {
	registerCollector(replicationSlotSubsystem, defaultEnabled, NewPGReplicationSlotCollector)
}

type PGReplicationSlotCollector struct {
	log log.Logger
}

func NewPGReplicationSlotCollector(config collectorConfig) (Collector, error) {
	return &PGReplicationSlotCollector{log: config.logger}, nil
}

var (
	pgReplicationSlotCurrentWalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_current_wal_lsn",
		),
		"current wal lsn value",
		[]string{"slot_name"}, nil,
	)
	pgReplicationSlotCurrentFlushDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_confirmed_flush_lsn",
		),
		"last lsn confirmed flushed to the replication slot",
		[]string{"slot_name"}, nil,
	)
	pgReplicationSlotIsActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSlotSubsystem,
			"slot_is_active",
		),
		"whether the replication slot is active or not",
		[]string{"slot_name"}, nil,
	)

	pgReplicationSlotQuery = `SELECT
		slot_name,
		pg_current_wal_lsn() - '0/0' AS current_wal_lsn,
		coalesce(confirmed_flush_lsn, '0/0') - '0/0',
		active
	FROM
		pg_replication_slots;`
)

func (PGReplicationSlotCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		pgReplicationSlotQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slotName string
		var walLSN int64
		var flushLSN int64
		var isActive bool
		if err := rows.Scan(&slotName, &walLSN, &flushLSN, &isActive); err != nil {
			return err
		}

		isActiveValue := 0
		if isActive {
			isActiveValue = 1
		}

		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotCurrentWalDesc,
			prometheus.GaugeValue, float64(walLSN), slotName,
		)
		if isActive {
			ch <- prometheus.MustNewConstMetric(
				pgReplicationSlotCurrentFlushDesc,
				prometheus.GaugeValue, float64(flushLSN), slotName,
			)
		}
		ch <- prometheus.MustNewConstMetric(
			pgReplicationSlotIsActiveDesc,
			prometheus.GaugeValue, float64(isActiveValue), slotName,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
