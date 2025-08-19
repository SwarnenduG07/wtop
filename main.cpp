#include "CLI11.hpp"
#include <iostream>
#include <memory>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>

#include "ui/display.hpp"
#include "telemetry/telemetry_manager.hpp"
#include "utils/config.hpp"
#include "utils/logger.hpp"

std::atomic<bool> g_running{true};

void signal_handler(int signal) {
    if (signal == SIGINT || signal == SIGTERM) {
        g_running = false;
    }
}

int main(int argc, char** argv) {
    CLI::App app{"wtop - Windows System Monitor", "wtop"};
    
    // CLI options
    int refresh_rate = 1000; // milliseconds
    bool enable_telemetry = true;
    std::string log_level = "info";
    std::string output_format = "interactive";
    bool show_processes = true;
    bool show_memory = true;
    bool show_cpu = true;
    bool show_network = true;
    bool show_disk = true;
    
    app.add_option("-r,--refresh", refresh_rate, "Refresh rate in milliseconds (default: 1000)")
        ->check(CLI::Range(100, 10000));
    
    app.add_flag("--no-telemetry", [&](size_t) { enable_telemetry = false; }, 
                 "Disable OpenTelemetry metrics collection");
    
    app.add_option("-l,--log-level", log_level, "Log level (debug, info, warn, error)")
        ->check(CLI::IsMember({"debug", "info", "warn", "error"}));
    
    app.add_option("-o,--output", output_format, "Output format (interactive, json, csv)")
        ->check(CLI::IsMember({"interactive", "json", "csv"}));
    
    app.add_flag("--no-processes", [&](size_t) { show_processes = false; }, 
                 "Hide process information");
    
    app.add_flag("--no-memory", [&](size_t) { show_memory = false; }, 
                 "Hide memory information");
    
    app.add_flag("--no-cpu", [&](size_t) { show_cpu = false; }, 
                 "Hide CPU information");
    
    app.add_flag("--no-network", [&](size_t) { show_network = false; }, 
                 "Hide network information");
    
    app.add_flag("--no-disk", [&](size_t) { show_disk = false; }, 
                 "Hide disk information");

    CLI11_PARSE(app, argc, argv);

    try {
        // Initialize logger
        wtop::utils::Logger::initialize(log_level);
        
        // Initialize configuration
        wtop::utils::Config config;
        config.refresh_rate = refresh_rate;
        config.enable_telemetry = enable_telemetry;
        config.output_format = output_format;
        config.show_processes = show_processes;
        config.show_memory = show_memory;
        config.show_cpu = show_cpu;
        config.show_network = show_network;
        config.show_disk = show_disk;
        
        // Initialize telemetry manager
        auto telemetry_manager = std::make_unique<wtop::telemetry::TelemetryManager>(config);
        if (enable_telemetry) {
            telemetry_manager->initialize();
        }
        
        // Initialize display
        auto display = std::make_unique<wtop::ui::Display>(config, std::move(telemetry_manager));
        
        // Set up signal handlers
        std::signal(SIGINT, signal_handler);
        std::signal(SIGTERM, signal_handler);
        
        // Main loop
        display->run(g_running);
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
