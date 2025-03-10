// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/spidernet-io/spiderpool/pkg/constant"
)

var (
	// meter is a global creator of metric instruments.
	meter api.Meter
	// globalEnableMetric determines whether to use metric or not
	globalEnableMetric bool
)

// InitMetricController will set up meter with the input param(required) and create a prometheus exporter.
// returns http handler and error
func InitMetricController(ctx context.Context, meterName string, enableMetric bool) (http.Handler, error) {
	if len(meterName) == 0 {
		return nil, fmt.Errorf("failed to init metric controller, meter name is asked to be set")
	}

	otelResource, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(constant.SpiderpoolAPIGroup),
		))
	if nil != err {
		return nil, err
	}

	exporter, err := prometheus.New()
	if nil != err {
		return nil, err
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(otelResource),
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Name: "*_histogram"},
			sdkmetric.Stream{Aggregation: aggregation.ExplicitBucketHistogram{
				Boundaries: []float64{0.1, 0.3, 0.5, 1, 3, 5, 7, 10, 15},
			}},
		)),
	)
	global.SetMeterProvider(provider)

	globalEnableMetric = enableMetric
	if globalEnableMetric {
		meter = global.Meter(meterName)
	} else {
		meter = api.NewNoopMeterProvider().Meter(meterName)
	}

	return promhttp.Handler(), nil
}

// NewMetricInt64Counter will create otel Int64Counter metric.
// The first param metricName is required and the second param is optional.
func NewMetricInt64Counter(metricName string, description string) (instrument.Int64Counter, error) {
	if len(metricName) == 0 {
		return nil, fmt.Errorf("failed to create metric Int64Counter, metric name is asked to be set")
	}
	return meter.Int64Counter(metricName, instrument.WithDescription(description))
}

// NewMetricFloat64Histogram will create otel Float64Histogram metric.
// The first param metricName is required and the second param is optional.
// Notice: if you want to match the quantile {0.1, 0.3, 0.5, 1, 3, 5, 7, 10, 15}, please let the metric name match regex "*_histogram",
// otherwise it will match the  otel default quantile.
func NewMetricFloat64Histogram(metricName string, description string) (instrument.Float64Histogram, error) {
	if len(metricName) == 0 {
		return nil, fmt.Errorf("failed to create metric Float64Histogram, metric name is asked to be set")
	}
	return meter.Float64Histogram(metricName, instrument.WithDescription(description))
}

// NewMetricFloat64Gauge will create otel Float64Gauge metric.
// The first param metricName is required and the second param is optional.
func NewMetricFloat64Gauge(metricName string, description string) (instrument.Float64ObservableGauge, error) {
	if len(metricName) == 0 {
		return nil, fmt.Errorf("failed to create metric Float64Guage, metric name is asked to be set")
	}

	return meter.Float64ObservableGauge(metricName, instrument.WithDescription(description))
}

// NewMetricInt64Gauge will create otel Int64Gauge metric.
// The first param metricName is required and the second param is optional.
func NewMetricInt64Gauge(metricName string, description string) (instrument.Int64ObservableGauge, error) {
	if len(metricName) == 0 {
		return nil, fmt.Errorf("failed to create metric Float64Guage, metric name is asked to be set")
	}

	return meter.Int64ObservableGauge(metricName, instrument.WithDescription(description))
}

var _ TimeRecorder = &timeRecorder{}

// timeRecorder owns a field to record start time.
type timeRecorder struct {
	startTime time.Time
}

// TimeRecorder will help you to compute time duration.
type TimeRecorder interface {
	SinceInSeconds() float64
}

// NewTimeRecorder will create TimeRecorder and record the current time.
func NewTimeRecorder() TimeRecorder {
	t := timeRecorder{}
	t.startTime = time.Now()
	return &t
}

// SinceInSeconds returns the duration of time since the start time as a float64.
func (t *timeRecorder) SinceInSeconds() float64 {
	return float64(time.Since(t.startTime)) / 1e9
}
