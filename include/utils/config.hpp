#pragma once

#include <string>

namespace wtop {
namespace utils {

struct Config {
    // Display settings
    int refresh_rate = 1000; // milliseconds
    std::string output_format = "interactive"; // interactive, json, csv
    
    // Feature toggles
    bool show_processes = true;
    bool show_memory = true;
    bool show_cpu = true;
    bool show_network = true;
    bool show_disk = true;
    
    // Telemetry settings
    bool enable_telemetry = true;
    std::string telemetry_endpoint = "http://localhost:4317";
    std::string service_name = "wtop";
    std::string service_version = "1.0.0";
    
    // Process settings
    int max_processes = 50;
    bool show_threads = false;
    
    // UI settings
    bool use_colors = true;
    bool show_help = true;
    
    // Performance settings
    int metric_buffer_size = 1000;
    int history_retention_seconds = 300; // 5 minutes
};

} // namespace utils
} // namespace wtop
