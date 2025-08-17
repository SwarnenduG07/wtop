#include "ui/display.hpp"
#include "utils/logger.hpp"
#include <iostream>
#include <iomanip>
#include <sstream>
#include <algorithm>
#include <conio.h>  // For Windows console input
#include <windows.h>

namespace wtop {
namespace ui {

class Display::TerminalManager {
public:
    TerminalManager() {
        // Get console handles
        h_stdout_ = GetStdHandle(STD_OUTPUT_HANDLE);
        h_stdin_ = GetStdHandle(STD_INPUT_HANDLE);
        
        // Save original console mode
        GetConsoleMode(h_stdin_, &original_mode_);
        
        // Set console to raw mode for better input handling
        DWORD new_mode = original_mode_ & ~(ENABLE_ECHO_INPUT | ENABLE_LINE_INPUT);
        SetConsoleMode(h_stdin_, new_mode);
        
        // Get console screen buffer info
        update_terminal_size();
        
        // Hide cursor
        CONSOLE_CURSOR_INFO cursor_info;
        GetConsoleCursorInfo(h_stdout_, &cursor_info);
        original_cursor_visible_ = cursor_info.bVisible;
        cursor_info.bVisible = FALSE;
        SetConsoleCursorInfo(h_stdout_, &cursor_info);
    }
    
    ~TerminalManager() {
        // Restore original console mode
        SetConsoleMode(h_stdin_, original_mode_);
        
        // Restore cursor visibility
        CONSOLE_CURSOR_INFO cursor_info;
        GetConsoleCursorInfo(h_stdout_, &cursor_info);
        cursor_info.bVisible = original_cursor_visible_;
        SetConsoleCursorInfo(h_stdout_, &cursor_info);
        
        // Clear screen and reset cursor
        clear_screen();
        set_cursor_position(0, 0);
    }
    
    void update_terminal_size() {
        CONSOLE_SCREEN_BUFFER_INFO csbi;
        if (GetConsoleScreenBufferInfo(h_stdout_, &csbi)) {
            width_ = csbi.srWindow.Right - csbi.srWindow.Left + 1;
            height_ = csbi.srWindow.Bottom - csbi.srWindow.Top + 1;
        }
    }
    
    void clear_screen() {
        COORD coord = {0, 0};
        DWORD written;
        CONSOLE_SCREEN_BUFFER_INFO csbi;
        
        GetConsoleScreenBufferInfo(h_stdout_, &csbi);
        FillConsoleOutputCharacterA(h_stdout_, ' ', csbi.dwSize.X * csbi.dwSize.Y, coord, &written);
        FillConsoleOutputAttribute(h_stdout_, csbi.wAttributes, csbi.dwSize.X * csbi.dwSize.Y, coord, &written);
        SetConsoleCursorPosition(h_stdout_, coord);
    }
    
    void set_cursor_position(int x, int y) {
        COORD coord = {static_cast<SHORT>(x), static_cast<SHORT>(y)};
        SetConsoleCursorPosition(h_stdout_, coord);
    }
    
    void set_color(int foreground, int background = 0) {
        SetConsoleTextAttribute(h_stdout_, foreground | (background << 4));
    }
    
    void reset_color() {
        SetConsoleTextAttribute(h_stdout_, 7); // Default white on black
    }
    
    int get_key() {
        if (_kbhit()) {
            return _getch();
        }
        return -1;
    }
    
    int width() const { return width_; }
    int height() const { return height_; }

private:
    HANDLE h_stdout_;
    HANDLE h_stdin_;
    DWORD original_mode_;
    BOOL original_cursor_visible_;
    int width_ = 80;
    int height_ = 24;
};

Display::Display(const utils::Config& config, 
                 std::unique_ptr<telemetry::TelemetryManager> telemetry_manager)
    : config_(config)
    , metrics_manager_(std::make_unique<metrics::MetricsManager>(config))
    , telemetry_manager_(std::move(telemetry_manager))
    , terminal_manager_(std::make_unique<TerminalManager>())
    , use_colors_(config.use_colors) {
}

Display::~Display() = default;

void Display::run(std::atomic<bool>& running) {
    initialize_display();
    
    try {
        // Start metrics collection
        metrics_manager_->start();
        
        utils::Logger::info("Starting wtop display loop");
        
        while (running && metrics_manager_->is_running()) {
            auto frame_start = std::chrono::steady_clock::now();
            
            // Handle input
            handle_input();
            
            // Render frame
            render_frame();
            
            // Calculate frame time and sleep
            auto frame_time = std::chrono::steady_clock::now() - frame_start;
            auto target_frame_time = std::chrono::milliseconds(config_.refresh_rate);
            
            if (frame_time < target_frame_time) {
                std::this_thread::sleep_for(target_frame_time - frame_time);
            }
        }
        
    } catch (const std::exception& e) {
        utils::Logger::error("Display loop error: {}", e.what());
    }
    
    cleanup_display();
}

void Display::initialize_display() {
    terminal_manager_->update_terminal_size();
    terminal_width_ = terminal_manager_->width();
    terminal_height_ = terminal_manager_->height();
    
    utils::Logger::info("Initialized display {}x{}", terminal_width_, terminal_height_);
}

void Display::cleanup_display() {
    terminal_manager_->clear_screen();
    terminal_manager_->set_cursor_position(0, 0);
    terminal_manager_->reset_color();
}

void Display::handle_input() {
    int key = terminal_manager_->get_key();
    if (key != -1) {
        process_key(key);
    }
}

void Display::render_frame() {
    if (config_.output_format == "json") {
        render_json();
        return;
    } else if (config_.output_format == "csv") {
        render_csv();
        return;
    }
    
    render_interactive();
}

void Display::render_interactive() {
    terminal_manager_->clear_screen();
    terminal_manager_->set_cursor_position(0, 0);
    
    render_header();
    
    if (show_help_) {
        render_help();
    } else {
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
    }
    
    render_footer();
}

void Display::render_header() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    if (use_colors_) {
        terminal_manager_->set_color(15); // Bright white
    }
    
    std::cout << "wtop - Windows System Monitor";
    
    // Show current time and uptime
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    
    std::cout << std::string(terminal_width_ - 50, ' ');
    std::cout << std::put_time(std::localtime(&time_t), "%H:%M:%S");
    std::cout << " up " << format_duration(std::chrono::seconds(static_cast<int>(metrics.system_uptime_seconds)));
    std::cout << std::endl;
    
    // Show basic system info
    std::cout << "Tasks: " << metrics.total_processes << " total, "
              << metrics.total_threads << " threads";
    
    std::cout << std::string(terminal_width_ - 40, ' ');
    std::cout << "Load: " << std::fixed << std::setprecision(2) << metrics.cpu.usage_percent << "%";
    std::cout << std::endl;
    
    if (use_colors_) {
        terminal_manager_->reset_color();
    }
    
    std::cout << std::string(terminal_width_, '-') << std::endl;
}

void Display::render_system_overview() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    // CPU information
    if (use_colors_) terminal_manager_->set_color(14); // Yellow
    std::cout << "CPU: " << metrics.cpu.name << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    std::cout << "Usage: " << format_percentage(metrics.cpu.usage_percent)
              << " (" << metrics.cpu.logical_processor_count << " cores)" << std::endl;
    
    // Memory information
    if (use_colors_) terminal_manager_->set_color(10); // Green
    std::cout << "Memory: ";
    if (use_colors_) terminal_manager_->reset_color();
    
    std::cout << format_bytes(metrics.memory.used_physical) << " / "
              << format_bytes(metrics.memory.total_physical) << " ("
              << format_percentage(metrics.memory.memory_load_percent) << ")" << std::endl;
    
    // Top processes by CPU
    if (use_colors_) terminal_manager_->set_color(12); // Red
    std::cout << "\nTop Processes by CPU:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    auto processes = metrics.processes;
    std::sort(processes.begin(), processes.end(),
              [](const metrics::ProcessInfo& a, const metrics::ProcessInfo& b) {
                  return a.cpu_percent > b.cpu_percent;
              });
    
    std::cout << std::left << std::setw(8) << "PID"
              << std::setw(20) << "Name"
              << std::setw(8) << "CPU%"
              << std::setw(12) << "Memory" << std::endl;
    
    for (size_t i = 0; i < std::min(size_t(10), processes.size()); ++i) {
        const auto& proc = processes[i];
        std::cout << std::left << std::setw(8) << proc.pid
                  << std::setw(20) << proc.name.substr(0, 19)
                  << std::setw(8) << std::fixed << std::setprecision(1) << proc.cpu_percent
                  << std::setw(12) << format_bytes(proc.memory_bytes) << std::endl;
    }
}

void Display::render_process_list() {
    auto metrics = metrics_manager_->get_latest_metrics();
    auto processes = metrics.processes;
    
    // Apply filter if set
    if (!process_filter_.empty()) {
        processes.erase(
            std::remove_if(processes.begin(), processes.end(),
                          [this](const metrics::ProcessInfo& proc) {
                              return proc.name.find(process_filter_) == std::string::npos;
                          }),
            processes.end());
    }
    
    // Sort processes
    sort_processes(processes);
    
    // Header
    if (use_colors_) terminal_manager_->set_color(11); // Cyan
    std::cout << std::left << std::setw(8) << "PID"
              << std::setw(25) << "Name"
              << std::setw(8) << "CPU%"
              << std::setw(12) << "Memory"
              << std::setw(8) << "Threads"
              << std::setw(10) << "Status" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    // Process list
    int max_processes = terminal_height_ - 8; // Reserve space for header/footer
    int start_idx = scroll_offset_;
    int end_idx = std::min(start_idx + max_processes, static_cast<int>(processes.size()));
    
    for (int i = start_idx; i < end_idx; ++i) {
        const auto& proc = processes[i];
        
        // Highlight high CPU usage
        if (use_colors_ && proc.cpu_percent > 80.0) {
            terminal_manager_->set_color(12); // Red
        } else if (use_colors_ && proc.cpu_percent > 50.0) {
            terminal_manager_->set_color(14); // Yellow
        }
        
        std::cout << std::left << std::setw(8) << proc.pid
                  << std::setw(25) << proc.name.substr(0, 24)
                  << std::setw(8) << std::fixed << std::setprecision(1) << proc.cpu_percent
                  << std::setw(12) << format_bytes(proc.memory_bytes)
                  << std::setw(8) << proc.thread_count
                  << std::setw(10) << proc.status.substr(0, 9) << std::endl;
        
        if (use_colors_) terminal_manager_->reset_color();
    }
}

void Display::render_memory_info() {
    auto metrics = metrics_manager_->get_latest_metrics();
    const auto& memory = metrics.memory;
    
    if (use_colors_) terminal_manager_->set_color(10); // Green
    std::cout << "Memory Information:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    std::cout << "Physical Memory:" << std::endl;
    std::cout << "  Total:     " << format_bytes(memory.total_physical) << std::endl;
    std::cout << "  Used:      " << format_bytes(memory.used_physical) << std::endl;
    std::cout << "  Available: " << format_bytes(memory.available_physical) << std::endl;
    std::cout << "  Usage:     " << format_percentage(memory.memory_load_percent) << std::endl;
    
    std::cout << "\nVirtual Memory:" << std::endl;
    std::cout << "  Total:     " << format_bytes(memory.total_virtual) << std::endl;
    std::cout << "  Used:      " << format_bytes(memory.used_virtual) << std::endl;
    std::cout << "  Available: " << format_bytes(memory.available_virtual) << std::endl;
    
    std::cout << "\nPage File:" << std::endl;
    std::cout << "  Total:     " << format_bytes(memory.total_page_file) << std::endl;
    std::cout << "  Used:      " << format_bytes(memory.used_page_file) << std::endl;
    std::cout << "  Available: " << format_bytes(memory.available_page_file) << std::endl;
}

void Display::render_cpu_info() {
    auto metrics = metrics_manager_->get_latest_metrics();
    const auto& cpu = metrics.cpu;
    
    if (use_colors_) terminal_manager_->set_color(14); // Yellow
    std::cout << "CPU Information:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    std::cout << "Name:              " << cpu.name << std::endl;
    std::cout << "Cores:             " << cpu.core_count << std::endl;
    std::cout << "Logical Processors: " << cpu.logical_processor_count << std::endl;
    std::cout << "Frequency:         " << cpu.frequency_mhz << " MHz" << std::endl;
    std::cout << "Overall Usage:     " << format_percentage(cpu.usage_percent) << std::endl;
    
    if (!cpu.per_core_usage.empty()) {
        std::cout << "\nPer-Core Usage:" << std::endl;
        for (size_t i = 0; i < cpu.per_core_usage.size(); ++i) {
            std::cout << "  Core " << i << ": " << format_percentage(cpu.per_core_usage[i]) << std::endl;
        }
    }
}

void Display::render_network_info() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    if (use_colors_) terminal_manager_->set_color(9); // Blue
    std::cout << "Network Interfaces:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    for (const auto& interface : metrics.network_interfaces) {
        std::cout << "\n" << interface.interface_name << " (" << interface.interface_description << "):" << std::endl;
        std::cout << "  Status:    " << (interface.is_connected ? "Connected" : "Disconnected") << std::endl;
        std::cout << "  IP:        " << interface.ip_address << std::endl;
        std::cout << "  Sent:      " << format_bytes(interface.bytes_sent) << std::endl;
        std::cout << "  Received:  " << format_bytes(interface.bytes_received) << std::endl;
        std::cout << "  Send Rate: " << format_rate(interface.send_rate_bps) << std::endl;
        std::cout << "  Recv Rate: " << format_rate(interface.receive_rate_bps) << std::endl;
    }
}

void Display::render_disk_info() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    if (use_colors_) terminal_manager_->set_color(13); // Magenta
    std::cout << "Disk Information:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    for (const auto& disk : metrics.disks) {
        std::cout << "\nDrive " << disk.drive_letter << ": (" << disk.volume_label << "):" << std::endl;
        std::cout << "  File System: " << disk.file_system << std::endl;
        std::cout << "  Total Space: " << format_bytes(disk.total_space) << std::endl;
        std::cout << "  Used Space:  " << format_bytes(disk.used_space) << std::endl;
        std::cout << "  Free Space:  " << format_bytes(disk.free_space) << std::endl;
        std::cout << "  Usage:       " << format_percentage(disk.usage_percent) << std::endl;
        std::cout << "  Read Rate:   " << format_rate(disk.read_rate_bps) << std::endl;
        std::cout << "  Write Rate:  " << format_rate(disk.write_rate_bps) << std::endl;
    }
}

void Display::render_help() {
    if (use_colors_) terminal_manager_->set_color(15); // Bright white
    std::cout << "wtop Help:" << std::endl;
    if (use_colors_) terminal_manager_->reset_color();
    
    std::cout << "\nNavigation:" << std::endl;
    std::cout << "  1-6    - Switch between views" << std::endl;
    std::cout << "  h, ?   - Show/hide this help" << std::endl;
    std::cout << "  q      - Quit" << std::endl;
    
    std::cout << "\nProcess List (View 2):" << std::endl;
    std::cout << "  p      - Sort by PID" << std::endl;
    std::cout << "  n      - Sort by Name" << std::endl;
    std::cout << "  c      - Sort by CPU" << std::endl;
    std::cout << "  m      - Sort by Memory" << std::endl;
    std::cout << "  t      - Sort by Threads" << std::endl;
    std::cout << "  r      - Reverse sort order" << std::endl;
    std::cout << "  /      - Filter processes" << std::endl;
    std::cout << "  Up/Dn  - Scroll process list" << std::endl;
    
    std::cout << "\nViews:" << std::endl;
    std::cout << "  1 - System Overview" << std::endl;
    std::cout << "  2 - Process List" << std::endl;
    std::cout << "  3 - Memory Information" << std::endl;
    std::cout << "  4 - CPU Information" << std::endl;
    std::cout << "  5 - Network Information" << std::endl;
    std::cout << "  6 - Disk Information" << std::endl;
}

void Display::render_footer() {
    // Move to bottom of screen
    terminal_manager_->set_cursor_position(0, terminal_height_ - 1);
    
    if (use_colors_) terminal_manager_->set_color(8); // Dark gray
    
    std::string footer = "Press 'h' for help, 'q' to quit";
    if (!process_filter_.empty()) {
        footer += " | Filter: " + process_filter_;
    }
    
    std::cout << footer;
    std::cout << std::string(terminal_width_ - footer.length(), ' ');
    
    if (use_colors_) terminal_manager_->reset_color();
}

void Display::render_json() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    // Simple JSON output (in a real implementation, use a JSON library)
    std::cout << "{\n";
    std::cout << "  \"timestamp\": \"" << std::chrono::duration_cast<std::chrono::seconds>(
        metrics.timestamp.time_since_epoch()).count() << "\",\n";
    std::cout << "  \"cpu_usage\": " << metrics.cpu.usage_percent << ",\n";
    std::cout << "  \"memory_usage_percent\": " << metrics.memory.memory_load_percent << ",\n";
    std::cout << "  \"total_processes\": " << metrics.total_processes << ",\n";
    std::cout << "  \"uptime_seconds\": " << metrics.system_uptime_seconds << "\n";
    std::cout << "}\n";
}

void Display::render_csv() {
    auto metrics = metrics_manager_->get_latest_metrics();
    
    // CSV header (print once)
    static bool header_printed = false;
    if (!header_printed) {
        std::cout << "timestamp,cpu_usage,memory_usage_percent,total_processes,uptime_seconds\n";
        header_printed = true;
    }
    
    std::cout << std::chrono::duration_cast<std::chrono::seconds>(
        metrics.timestamp.time_since_epoch()).count() << ","
              << metrics.cpu.usage_percent << ","
              << metrics.memory.memory_load_percent << ","
              << metrics.total_processes << ","
              << metrics.system_uptime_seconds << "\n";
}

void Display::process_key(int key) {
    switch (key) {
        case 'q':
        case 'Q':
        case 27: // ESC
            // Signal to quit (handled by main loop)
            break;
            
        case 'h':
        case 'H':
        case '?':
            show_help_ = !show_help_;
            break;
            
        case '1':
            change_display_mode(DisplayMode::OVERVIEW);
            break;
        case '2':
            change_display_mode(DisplayMode::PROCESSES);
            break;
        case '3':
            change_display_mode(DisplayMode::MEMORY);
            break;
        case '4':
            change_display_mode(DisplayMode::CPU);
            break;
        case '5':
            change_display_mode(DisplayMode::NETWORK);
            break;
        case '6':
            change_display_mode(DisplayMode::DISK);
            break;
            
        // Process list controls
        case 'p':
        case 'P':
            toggle_sort_column(SortColumn::PID);
            break;
        case 'n':
        case 'N':
            toggle_sort_column(SortColumn::NAME);
            break;
        case 'c':
        case 'C':
            toggle_sort_column(SortColumn::CPU);
            break;
        case 'm':
        case 'M':
            toggle_sort_column(SortColumn::MEMORY);
            break;
        case 't':
        case 'T':
            toggle_sort_column(SortColumn::THREADS);
            break;
        case 'r':
        case 'R':
            sort_order_ = (sort_order_ == SortOrder::ASCENDING) ? 
                         SortOrder::DESCENDING : SortOrder::ASCENDING;
            break;
            
        // Scrolling
        case 72: // Up arrow (Windows)
            if (scroll_offset_ > 0) scroll_offset_--;
            break;
        case 80: // Down arrow (Windows)
            scroll_offset_++;
            break;
    }
}

void Display::toggle_sort_column(SortColumn column) {
    if (sort_column_ == column) {
        sort_order_ = (sort_order_ == SortOrder::ASCENDING) ? 
                     SortOrder::DESCENDING : SortOrder::ASCENDING;
    } else {
        sort_column_ = column;
        sort_order_ = SortOrder::DESCENDING;
    }
    scroll_offset_ = 0; // Reset scroll when changing sort
}

void Display::change_display_mode(DisplayMode mode) {
    current_mode_ = mode;
    scroll_offset_ = 0;
    show_help_ = false;
}

void Display::sort_processes(std::vector<metrics::ProcessInfo>& processes) {
    std::sort(processes.begin(), processes.end(),
              [this](const metrics::ProcessInfo& a, const metrics::ProcessInfo& b) {
                  bool result = false;
                  
                  switch (sort_column_) {
                      case SortColumn::PID:
                          result = a.pid < b.pid;
                          break;
                      case SortColumn::NAME:
                          result = a.name < b.name;
                          break;
                      case SortColumn::CPU:
                          result = a.cpu_percent < b.cpu_percent;
                          break;
                      case SortColumn::MEMORY:
                          result = a.memory_bytes < b.memory_bytes;
                          break;
                      case SortColumn::THREADS:
                          result = a.thread_count < b.thread_count;
                          break;
                  }
                  
                  return (sort_order_ == SortOrder::ASCENDING) ? result : !result;
              });
}

std::string Display::format_bytes(uint64_t bytes) {
    const char* units[] = {"B", "KB", "MB", "GB", "TB"};
    int unit = 0;
    double size = static_cast<double>(bytes);
    
    while (size >= 1024.0 && unit < 4) {
        size /= 1024.0;
        unit++;
    }
    
    std::ostringstream oss;
    oss << std::fixed << std::setprecision(1) << size << " " << units[unit];
    return oss.str();
}

std::string Display::format_percentage(double percent) {
    std::ostringstream oss;
    oss << std::fixed << std::setprecision(1) << percent << "%";
    return oss.str();
}

std::string Display::format_duration(std::chrono::seconds duration) {
    auto hours = std::chrono::duration_cast<std::chrono::hours>(duration);
    auto minutes = std::chrono::duration_cast<std::chrono::minutes>(duration % std::chrono::hours(1));
    
    std::ostringstream oss;
    oss << hours.count() << ":" << std::setfill('0') << std::setw(2) << minutes.count();
    return oss.str();
}

std::string Display::format_rate(uint64_t rate_bps) {
    return format_bytes(rate_bps) + "/s";
}

} // namespace ui
} // namespace wtop
