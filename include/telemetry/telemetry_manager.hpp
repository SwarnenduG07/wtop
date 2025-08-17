#pragma once

#include "utils/config.hpp"
#include "metrics/system_metrics.hpp"
#include <memory>
#include <string>

// Forward declarations for OpenTelemetry types
namespace opentelemetry {
namespace metrics {
class MeterProvider;
class Meter;
} // namespace metrics
namespace trace {
class TracerProvider;
class Tracer;
} // namespace trace
} // namespace opentelemetry

namespace wtop {
namespace telemetry {

class TelemetryManager {
public:
    explicit TelemetryManager(const utils::Config& config);
    ~TelemetryManager();
    
    void initialize();
    void shutdown();
    
    void record_system_metrics(const metrics::SystemMetrics& metrics);
    void record_process_metrics(const std::vector<metrics::ProcessInfo>& processes);
    void record_memory_metrics(const metrics::MemoryInfo& memory);
    void record_cpu_metrics(const metrics::CpuInfo& cpu);
    void record_network_metrics(const std::vector<metrics::NetworkInfo>& network);
    void record_disk_metrics(const std::vector<metrics::DiskInfo>& disks);
    
    // Tracing support
    void start_span(const std::string& name);
    void end_span();
    void add_span_attribute(const std::string& key, const std::string& value);
    void add_span_attribute(const std::string& key, int64_t value);
    void add_span_attribute(const std::string& key, double value);
    
    bool is_initialized() const { return initialized_; }

private:
    void setup_meter_provider();
    void setup_tracer_provider();
    void create_instruments();
    
    const utils::Config& config_;
    bool initialized_ = false;
    
    // OpenTelemetry components
    std::shared_ptr<opentelemetry::metrics::MeterProvider> meter_provider_;
    std::shared_ptr<opentelemetry::metrics::Meter> meter_;
    std::shared_ptr<opentelemetry::trace::TracerProvider> tracer_provider_;
    std::shared_ptr<opentelemetry::trace::Tracer> tracer_;
    
    // Metric instruments - will be defined in implementation
    class Instruments;
    std::unique_ptr<Instruments> instruments_;
};

} // namespace telemetry
} // namespace wtop
