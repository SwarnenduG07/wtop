package main

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/disk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.16.0"
)

func initMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("wtop-go")),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(1*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func collectMetrics(ctx context.Context) {
	meter := otel.Meter("wtop-go")

	// CPU metrics
	cpuGauge, _ := meter.Float64ObservableGauge("system.cpu.usage")
	meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		percent, _ := cpu.Percent(0, false)
		if len(percent) > 0 {
			log.Printf("CPU: %.2f%%", percent[0])
			o.ObserveFloat64(cpuGauge, percent[0])
		}
		return nil
	}, cpuGauge)

	// Memory metrics
	memUsage, _ := meter.Int64ObservableGauge("system.memory.usage")
	memAvailable, _ := meter.Int64ObservableGauge("system.memory.available")
	meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		v, _ := mem.VirtualMemory()
		log.Printf("Memory: Used=%dMB, Available=%dMB", v.Used/1024/1024, v.Available/1024/1024)
		o.ObserveInt64(memUsage, int64(v.Used))
		o.ObserveInt64(memAvailable, int64(v.Available))
		return nil
	}, memUsage, memAvailable)

	// Network metrics
	netBytes, _ := meter.Int64ObservableCounter("system.network.bytes")
	meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats, _ := net.IOCounters(false)
		if len(stats) > 0 {
			o.ObserveInt64(netBytes, int64(stats[0].BytesSent+stats[0].BytesRecv))
		}
		return nil
	}, netBytes)

	// Disk metrics
	diskFree, _ := meter.Int64ObservableGauge("system.disk.free")
	meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		usage, _ := disk.Usage("C:")
		o.ObserveInt64(diskFree, int64(usage.Free))
		return nil
	}, diskFree)
}

func main() {
	ctx := context.Background()

	mp, err := initMeterProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer mp.Shutdown(ctx)

	collectMetrics(ctx)

	log.Println("wtop-go started. Collecting metrics...")
	select {} // Keep running
}
