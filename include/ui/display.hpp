#pragma once

#include "utils/config.hpp"
#include "metrics/metrics_manager.hpp"
#include "telemetry/telemetry_manager.hpp"
#include <memory>
#include <atomic>
#include <string>

namespace wtop {
namespace ui {

enum class DisplayMode {
    PROCESSES,
    MEMORY,
    CPU,
    NETWORK,
    DISK,
    OVERVIEW
};

enum class SortColumn {
    PID,
    NAME,
    CPU,
    MEMORY,
    THREADS
};

enum class SortOrder {
    ASCENDING,
    DESCENDING
};

class Display {
public:
    Display(const utils::Config& config, 
            std::unique_ptr<telemetry::TelemetryManager> telemetry_manager);
    ~Display();
    
    void run(std::atomic<bool>& running);
    
private:
    void initialize_display();
    void cleanup_display();
    void handle_input();
    void render_frame();
    
    // Rendering methods
    void render_header();
    void render_system_overview();
    void render_process_list();
    void render_memory_info();
    void render_cpu_info();
    void render_network_info();
    void render_disk_info();
    void render_help();
    void render_footer();
    
    // Interactive mode rendering
    void render_interactive();
    
    // Non-interactive output formats
    void render_json();
    void render_csv();
    
    // Input handling
    void process_key(int key);
    void toggle_sort_column(SortColumn column);
    void change_display_mode(DisplayMode mode);
    void filter_processes(const std::string& filter);
    
    // Utility methods
    std::string format_bytes(uint64_t bytes);
    std::string format_percentage(double percent);
    std::string format_duration(std::chrono::seconds duration);
    std::string format_rate(uint64_t rate_bps);
    void sort_processes(std::vector<metrics::ProcessInfo>& processes);
    
    const utils::Config& config_;
    std::unique_ptr<metrics::MetricsManager> metrics_manager_;
    std::unique_ptr<telemetry::TelemetryManager> telemetry_manager_;
    
    // Display state
    DisplayMode current_mode_ = DisplayMode::OVERVIEW;
    SortColumn sort_column_ = SortColumn::CPU;
    SortOrder sort_order_ = SortOrder::DESCENDING;
    std::string process_filter_;
    bool show_help_ = false;
    int scroll_offset_ = 0;
    
    // Terminal dimensions
    int terminal_width_ = 80;
    int terminal_height_ = 24;
    
    // Colors and formatting
    bool use_colors_ = true;
    
    class TerminalManager;
    std::unique_ptr<TerminalManager> terminal_manager_;
};

} // namespace ui
} // namespace wtop
