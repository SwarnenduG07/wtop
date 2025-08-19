#include "ui/display.hpp"
#include <iostream>
#include <thread>
#include <chrono>

namespace wtop {
namespace ui {

// Forward declaration for the TerminalManager class
class Display::TerminalManager {
public:
    TerminalManager() = default;
    ~TerminalManager() = default;
    
    void initialize() {
        std::cout << "Terminal initialized" << std::endl;
    }
    
    void cleanup() {
        std::cout << "Terminal cleaned up" << std::endl;
    }
};

Display::Display(const utils::Config& config, std::unique_ptr<telemetry::TelemetryManager> telemetry_manager) 
    : config_(config), 
      telemetry_manager_(std::move(telemetry_manager)),
      metrics_manager_(std::make_unique<metrics::MetricsManager>(config)),
      terminal_manager_(std::make_unique<TerminalManager>()) {
    std::cout << "Display created" << std::endl;
}

Display::~Display() {
    cleanup_display();
    std::cout << "Display destroyed" << std::endl;
}

void Display::run(std::atomic<bool>& running) {
    initialize_display();
    
    std::cout << "Display running..." << std::endl;
    
    metrics_manager_->start();
    
    while (running.load()) {
        render_frame();
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
    }
    
    metrics_manager_->stop();
    cleanup_display();
    
    std::cout << "Display stopped" << std::endl;
}

void Display::initialize_display() {
    terminal_manager_->initialize();
    std::cout << "Display initialized" << std::endl;
}

void Display::cleanup_display() {
    terminal_manager_->cleanup();
    std::cout << "Display cleaned up" << std::endl;
}

void Display::handle_input() {
    // Stub for input handling
}

void Display::render_frame() {
    if (config_.output_format == "json") {
        render_json();
    } else if (config_.output_format == "csv") {
        render_csv();
    } else {
        render_interactive();
    }
}

void Display::render_interactive() {
    render_header();
    
    switch (current_mode_) {
        case DisplayMode::OVERVIEW:
            render_system_overview();
            break;
        case DisplayMode::PROCESSES:
            render_process_list();
            break;
        case DisplayMode::MEMORY:
            render_memory_info();
            break;
        case DisplayMode::CPU:
            render_cpu_info();
            break;
        case DisplayMode::NETWORK:
            render_network_info();
            break;
        case DisplayMode::DISK:
            render_disk_info();
            break;
    }
    
    if (show_help_) {
        render_help();
    }
    
    render_footer();
}

void Display::render_header() {
    std::cout << "wtop - Windows System Monitor" << std::endl;
}

void Display::render_system_overview() {
    auto metrics = metrics_manager_->get_latest_metrics();
    std::cout << "System Overview - Processes: " << metrics.total_processes 
              << ", Threads: " << metrics.total_threads << std::endl;
}

void Display::render_process_list() {
    std::cout << "Process List" << std::endl;
}

void Display::render_memory_info() {
    std::cout << "Memory Information" << std::endl;
}

void Display::render_cpu_info() {
    std::cout << "CPU Information" << std::endl;
}

void Display::render_network_info() {
    std::cout << "Network Information" << std::endl;
}

void Display::render_disk_info() {
    std::cout << "Disk Information" << std::endl;
}

void Display::render_help() {
    std::cout << "Help: Press 'q' to quit" << std::endl;
}

void Display::render_footer() {
    std::cout << "Footer" << std::endl;
}

void Display::render_json() {
    std::cout << "{\"status\": \"monitoring\"}" << std::endl;
}

void Display::render_csv() {
    std::cout << "timestamp,processes,threads" << std::endl;
    auto metrics = metrics_manager_->get_latest_metrics();
    std::cout << "now," << metrics.total_processes << "," << metrics.total_threads << std::endl;
}

void Display::process_key(int key) {
    // Stub for key processing
}

void Display::toggle_sort_column(SortColumn column) {
    // Stub for sort column toggle
}

void Display::change_display_mode(DisplayMode mode) {
    current_mode_ = mode;
}

void Display::filter_processes(const std::string& filter) {
    process_filter_ = filter;
}

std::string Display::format_bytes(uint64_t bytes) {
    return std::to_string(bytes) + " B";
}

std::string Display::format_percentage(double percent) {
    return std::to_string(percent) + "%";
}

std::string Display::format_duration(std::chrono::seconds duration) {
    return std::to_string(duration.count()) + "s";
}

std::string Display::format_rate(uint64_t rate_bps) {
    return std::to_string(rate_bps) + " bps";
}

void Display::sort_processes(std::vector<metrics::ProcessInfo>& processes) {
    // Stub for process sorting
}

} // namespace ui
} // namespace wtop
