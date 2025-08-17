#pragma once

#include <vector>
#include <string>
#include <memory>
#include <chrono>

namespace wtop {
namespace metrics {

struct ProcessInfo {
    uint32_t pid;
    std::string name;
    std::string command_line;
    double cpu_percent;
    uint64_t memory_bytes;
    uint64_t virtual_memory_bytes;
    uint32_t thread_count;
    std::string status;
    std::chrono::system_clock::time_point start_time;
    std::string user;
};

struct MemoryInfo {
    uint64_t total_physical;
    uint64_t available_physical;
    uint64_t used_physical;
    uint64_t total_virtual;
    uint64_t available_virtual;
    uint64_t used_virtual;
    uint64_t total_page_file;
    uint64_t available_page_file;
    uint64_t used_page_file;
    double memory_load_percent;
};

struct CpuInfo {
    std::string name;
    uint32_t core_count;
    uint32_t logical_processor_count;
    double usage_percent;
    std::vector<double> per_core_usage;
    uint64_t frequency_mhz;
    double temperature_celsius;
};

struct NetworkInfo {
    std::string interface_name;
    std::string interface_description;
    uint64_t bytes_sent;
    uint64_t bytes_received;
    uint64_t packets_sent;
    uint64_t packets_received;
    uint64_t send_rate_bps;
    uint64_t receive_rate_bps;
    bool is_connected;
    std::string ip_address;
};

struct DiskInfo {
    std::string drive_letter;
    std::string file_system;
    std::string volume_label;
    uint64_t total_space;
    uint64_t free_space;
    uint64_t used_space;
    double usage_percent;
    uint64_t read_rate_bps;
    uint64_t write_rate_bps;
    uint32_t read_iops;
    uint32_t write_iops;
};

struct SystemMetrics {
    std::chrono::system_clock::time_point timestamp;
    std::vector<ProcessInfo> processes;
    MemoryInfo memory;
    CpuInfo cpu;
    std::vector<NetworkInfo> network_interfaces;
    std::vector<DiskInfo> disks;
    double system_uptime_seconds;
    uint32_t total_processes;
    uint32_t total_threads;
};

class SystemMetricsCollector {
public:
    SystemMetricsCollector();
    ~SystemMetricsCollector();
    
    SystemMetrics collect();
    
    // Individual metric collection methods
    std::vector<ProcessInfo> collect_process_info();
    MemoryInfo collect_memory_info();
    CpuInfo collect_cpu_info();
    std::vector<NetworkInfo> collect_network_info();
    std::vector<DiskInfo> collect_disk_info();
    
private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
};

} // namespace metrics
} // namespace wtop
