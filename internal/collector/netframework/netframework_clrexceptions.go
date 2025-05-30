// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

//go:build windows

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrExceptions() {
	c.numberOfExceptionsThrown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_thrown_total"),
		"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
		[]string{"process"},
		nil,
	)
	c.numberOfFilters = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_filters_total"),
		"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
		[]string{"process"},
		nil,
	)
	c.numberOfFinally = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_finallys_total"),
		"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
		[]string{"process"},
		nil,
	)
	c.throwToCatchDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_throw_to_catch_depth_total"),
		"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
		[]string{"process"},
		nil,
	)
}

type Win32_PerfRawData_NETFramework_NETCLRExceptions struct {
	Name string `mi:"Name"`

	NumberofExcepsThrown       uint32 `mi:"NumberofExcepsThrown"`
	NumberofExcepsThrownPersec uint32 `mi:"NumberofExcepsThrownPersec"`
	NumberofFiltersPersec      uint32 `mi:"NumberofFiltersPersec"`
	NumberofFinallysPersec     uint32 `mi:"NumberofFinallysPersec"`
	ThrowToCatchDepthPersec    uint32 `mi:"ThrowToCatchDepthPersec"`
}

func (c *Collector) collectClrExceptions(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRExceptions
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRExceptions"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfExceptionsThrown,
			prometheus.CounterValue,
			float64(process.NumberofExcepsThrown),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFilters,
			prometheus.CounterValue,
			float64(process.NumberofFiltersPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFinally,
			prometheus.CounterValue,
			float64(process.NumberofFinallysPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.throwToCatchDepth,
			prometheus.CounterValue,
			float64(process.ThrowToCatchDepthPersec),
			process.Name,
		)
	}

	return nil
}
