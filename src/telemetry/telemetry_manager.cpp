#include "telemetry/telemetry_manager.hpp"
#include "utils/logger.hpp"

// OpenTelemetry includes
#include <opentelemetry/exporters/otlp/otlp_grpc_exporter_factory.h>
#include <opentelemetry/exporters/otlp/otlp_grpc_metric_exporter_factory.h>
#include <opentelemetry/metrics/provider.h>
#include <opentelemetry/sdk/metrics/meter_provider.h>
#include <opentelemetry/sdk/metrics/metric_reader.h>
#include <opentelemetry/sdk/metrics/periodic_exporting_metric_reader.h>
#include <opentelemetry/sdk/trace/tracer_provider.h>
#include <opentelemetry/sdk/trace/batch_span_processor.h>
#include <opentelemetry/trace/provider.h>
#include <opentelemetry/sdk/resource/resource.h>
#include <opentelemetry/sdk/common/attribute_utils.h>

namespace wtop {
namespace telemetry {

class TelemetryManager::Instruments {
public:
    // System metrics
    std::unique_ptr<opentelemetry::metrics::Histogram<double>> cpu_usage_histogram;
    std::unique_ptr<opentelemetry::metrics::Histogram<uint64_t>> memory_usage_histogram;
    std::unique_ptr<opentelemetry::metrics::Counter<uint64_t>> network_bytes_counter;
    std::unique_ptr<opentelemetry::metrics::Counter<uint64_t>> disk_io_counter;
    
    // Process metrics
    std::unique_ptr<opentelemetry::metrics::UpDownCounter<int64_t>> process_count_gauge;
    std::unique_ptr<opentelemetry::metrics::Histogram<double>> process_cpu_histogram;
    std::unique_ptr<opentelemetry::metrics::Histogram<uint64_t>> process_memory_histogram;
    
    // System gauges
    std::unique_ptr<opentelemetry::metrics::UpDownCounter<int64_t>> memory_available_gauge;
    std::unique_ptr<opentelemetry::metrics::UpDownCounter<int64_t>> disk_free_space_gauge;
};

TelemetryManager::TelemetryManager(const utils::Config& config)
    : config_(config)
    , instruments_(std::make_unique<Instruments>()) {
}

TelemetryManager::~TelemetryManager() {
    shutdown();
}

void TelemetryManager::initialize() {
    if (initialized_) {
        return;
    }
    
    try {
        utils::Logger::info("Initializing OpenTelemetry");
        
        setup_meter_provider();
        setup_tracer_provider();
        create_instruments();
        
        initialized_ = true;
        utils::Logger::info("OpenTelemetry initialized successfully");
        
    } catch (const std::exception& e) {
        utils::Logger::error("Failed to initialize OpenTelemetry: {}", e.what());
        throw;
    }
}

void TelemetryManager::shutdown() {
    if (!initialized_) {
        return;
    }
    
    utils::Logger::info("Shutting down OpenTelemetry");
    
    // Reset instruments
    instruments_ = std::make_unique<Instruments>();
    
    // Shutdown providers
    meter_.reset();
    meter_provider_.reset();
    tracer_.reset();
    tracer_provider_.reset();
    
    initialized_ = false;
}

void TelemetryManager::setup_meter_provider() {
    // Create resource with service information
    auto resource = opentelemetry::sdk::resource::Resource::Create({
        {"service.name", config_.service_name},
        {"service.version", config_.service_version},
        {"telemetry.sdk.name", "opentelemetry"},
        {"telemetry.sdk.language", "cpp"},
        {"telemetry.sdk.version", OPENTELEMETRY_VERSION}
    });
    
    // Create OTLP exporter
    opentelemetry::exporter::otlp::OtlpGrpcMetricExporterOptions exporter_options;
    exporter_options.endpoint = config_.telemetry_endpoint;
    
    auto exporter = opentelemetry::exporter::otlp::OtlpGrpcMetricExporterFactory::Create(exporter_options);
    
    // Create periodic reader
    opentelemetry::sdk::metrics::PeriodicExportingMetricReaderOptions reader_options;
    reader_options.export_interval_millis = std::chrono::milliseconds(config_.refresh_rate);
    reader_options.export_timeout_millis = std::chrono::milliseconds(config_.refresh_rate / 2);
    
    auto reader = std::unique_ptr<opentelemetry::sdk::metrics::MetricReader>(
        new opentelemetry::sdk::metrics::PeriodicExportingMetricReader(
            std::move(exporter), reader_options));
    
    // Create meter provider
    auto provider = std::unique_ptr<opentelemetry::sdk::metrics::MeterProvider>(
        new opentelemetry::sdk::metrics::MeterProvider(
            std::move(resource), {std::move(reader)}));
    
    meter_provider_ = std::move(provider);
    
    // Set global meter provider
    opentelemetry::metrics::Provider::SetMeterProvider(meter_provider_);
    
    // Get meter
    meter_ = meter_provider_->GetMeter("wtop", "1.0.0");
}

void TelemetryManager::setup_tracer_provider() {
    // Create resource
    auto resource = opentelemetry::sdk::resource::Resource::Create({
        {"service.name", config_.service_name},
        {"service.version", config_.service_version}
    });
    
    // Create OTLP exporter for traces
    opentelemetry::exporter::otlp::OtlpGrpcExporterOptions exporter_options;
    exporter_options.endpoint = config_.telemetry_endpoint;
    
    auto exporter = opentelemetry::exporter::otlp::OtlpGrpcExporterFactory::Create(exporter_options);
    
    // Create batch span processor
    opentelemetry::sdk::trace::BatchSpanProcessorOptions processor_options;
    auto processor = std::unique_ptr<opentelemetry::sdk::trace::SpanProcessor>(
        new opentelemetry::sdk::trace::BatchSpanProcessor(std::move(exporter), processor_options));
    
    // Create tracer provider
    auto provider = std::unique_ptr<opentelemetry::sdk::trace::TracerProvider>(
        new opentelemetry::sdk::trace::TracerProvider(
            std::move(processor), resource));
    
    tracer_provider_ = std::move(provider);
    
    // Set global tracer provider
    opentelemetry::trace::Provider::SetTracerProvider(tracer_provider_);
    
    // Get tracer
    tracer_ = tracer_provider_->GetTracer("wtop", "1.0.0");
}

void TelemetryManager::create_instruments() {
    if (!meter_) {
        throw std::runtime_error("Meter not initialized");
    }
    
    // Create histograms for distributions
    instruments_->cpu_usage_histogram = meter_->CreateDoubleHistogram(
        "system.cpu.usage", "CPU usage percentage", "%");
    
    instruments_->memory_usage_histogram = meter_->CreateUInt64Histogram(
        "system.memory.usage", "Memory usage in bytes", "bytes");
    
    instruments_->process_cpu_histogram = meter_->CreateDoubleHistogram(
        "process.cpu.usage", "Process CPU usage percentage", "%");
    
    instruments_->process_memory_histogram = meter_->CreateUInt64Histogram(
        "process.memory.usage", "Process memory usage in bytes", "bytes");
    
    // Create counters for cumulative metrics
    instruments_->network_bytes_counter = meter_->CreateUInt64Counter(
        "system.network.bytes", "Network bytes transferred", "bytes");
    
    instruments_->disk_io_counter = meter_->CreateUInt64Counter(
        "system.disk.io", "Disk I/O operations", "operations");
    
    // Create gauges for current values
    instruments_->process_count_gauge = meter_->CreateInt64UpDownCounter(
        "system.process.count", "Number of running processes", "processes");
    
    instruments_->memory_available_gauge = meter_->CreateInt64UpDownCounter(
        "system.memory.available", "Available memory in bytes", "bytes");
    
    instruments_->disk_free_space_gauge = meter_->CreateInt64UpDownCounter(
        "system.disk.free", "Free disk space in bytes", "bytes");
}

void TelemetryManager::record_system_metrics(const metrics::SystemMetrics& metrics) {
    if (!initialized_) {
        return;
    }
    
    try {
        record_cpu_metrics(metrics.cpu);
        record_memory_metrics(metrics.memory);
        record_process_metrics(metrics.processes);
        record_network_metrics(metrics.network_interfaces);
        record_disk_metrics(metrics.disks);
        
        // Record system-level metrics
        if (instruments_->process_count_gauge) {
            instruments_->process_count_gauge->Add(static_cast<int64_t>(metrics.total_processes));
        }
        
    } catch (const std::exception& e) {
        utils::Logger::error("Failed to record system metrics: {}", e.what());
    }
}

void TelemetryManager::record_process_metrics(const std::vector<metrics::ProcessInfo>& processes) {
    if (!initialized_ || !instruments_->process_cpu_histogram || !instruments_->process_memory_histogram) {
        return;
    }
    
    for (const auto& process : processes) {
        auto attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
            {"process.name", process.name},
            {"process.pid", std::to_string(process.pid)}
        });
        
        instruments_->process_cpu_histogram->Record(process.cpu_percent, attributes);
        instruments_->process_memory_histogram->Record(process.memory_bytes, attributes);
    }
}

void TelemetryManager::record_memory_metrics(const metrics::MemoryInfo& memory) {
    if (!initialized_) {
        return;
    }
    
    if (instruments_->memory_usage_histogram) {
        instruments_->memory_usage_histogram->Record(memory.used_physical);
    }
    
    if (instruments_->memory_available_gauge) {
        instruments_->memory_available_gauge->Add(static_cast<int64_t>(memory.available_physical));
    }
}

void TelemetryManager::record_cpu_metrics(const metrics::CpuInfo& cpu) {
    if (!initialized_ || !instruments_->cpu_usage_histogram) {
        return;
    }
    
    instruments_->cpu_usage_histogram->Record(cpu.usage_percent);
    
    // Record per-core metrics
    for (size_t i = 0; i < cpu.per_core_usage.size(); ++i) {
        auto attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
            {"cpu.core", std::to_string(i)}
        });
        instruments_->cpu_usage_histogram->Record(cpu.per_core_usage[i], attributes);
    }
}

void TelemetryManager::record_network_metrics(const std::vector<metrics::NetworkInfo>& network) {
    if (!initialized_ || !instruments_->network_bytes_counter) {
        return;
    }
    
    for (const auto& interface : network) {
        auto sent_attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
            {"network.interface", interface.interface_name},
            {"network.direction", "sent"}
        });
        
        auto received_attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
            {"network.interface", interface.interface_name},
            {"network.direction", "received"}
        });
        
        instruments_->network_bytes_counter->Add(interface.bytes_sent, sent_attributes);
        instruments_->network_bytes_counter->Add(interface.bytes_received, received_attributes);
    }
}

void TelemetryManager::record_disk_metrics(const std::vector<metrics::DiskInfo>& disks) {
    if (!initialized_) {
        return;
    }
    
    for (const auto& disk : disks) {
        auto attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
            {"disk.drive", disk.drive_letter}
        });
        
        if (instruments_->disk_free_space_gauge) {
            instruments_->disk_free_space_gauge->Add(static_cast<int64_t>(disk.free_space), attributes);
        }
        
        if (instruments_->disk_io_counter) {
            auto read_attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
                {"disk.drive", disk.drive_letter},
                {"disk.operation", "read"}
            });
            
            auto write_attributes = opentelemetry::common::KeyValueIterableView<std::map<std::string, std::string>>({
                {"disk.drive", disk.drive_letter},
                {"disk.operation", "write"}
            });
            
            instruments_->disk_io_counter->Add(disk.read_iops, read_attributes);
            instruments_->disk_io_counter->Add(disk.write_iops, write_attributes);
        }
    }
}

void TelemetryManager::start_span(const std::string& name) {
    if (!initialized_ || !tracer_) {
        return;
    }
    
    // This is a simplified implementation
    // In a real implementation, you'd want to manage span context properly
    auto span = tracer_->StartSpan(name);
}

void TelemetryManager::end_span() {
    // Implementation would depend on how spans are managed
}

void TelemetryManager::add_span_attribute(const std::string& key, const std::string& value) {
    // Implementation would depend on current span context
}

void TelemetryManager::add_span_attribute(const std::string& key, int64_t value) {
    // Implementation would depend on current span context
}

void TelemetryManager::add_span_attribute(const std::string& key, double value) {
    // Implementation would depend on current span context
}

} // namespace telemetry
} // namespace wtop
