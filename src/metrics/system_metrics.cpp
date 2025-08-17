#include "metrics/system_metrics.hpp"
#include "utils/logger.hpp"

#ifdef _WIN32
#include <windows.h>
#include <psapi.h>
#include <pdh.h>
#include <pdhmsg.h>
#include <iphlpapi.h>
#include <tlhelp32.h>
#include <winternl.h>
#pragma comment(lib, "pdh.lib")
#pragma comment(lib, "iphlpapi.lib")
#pragma comment(lib, "psapi.lib")
#endif

#include <algorithm>
#include <sstream>
#include <iomanip>

namespace wtop {
namespace metrics {

class SystemMetricsCollector::Impl {
public:
    Impl() {
#ifdef _WIN32
        initialize_pdh();
#endif
    }
    
    ~Impl() {
#ifdef _WIN32
        cleanup_pdh();
#endif
    }
    
    SystemMetrics collect() {
        SystemMetrics metrics;
        metrics.timestamp = std::chrono::system_clock::now();
        
        try {
            metrics.processes = collect_process_info();
            metrics.memory = collect_memory_info();
            metrics.cpu = collect_cpu_info();
            metrics.network_interfaces = collect_network_info();
            metrics.disks = collect_disk_info();
            metrics.system_uptime_seconds = get_system_uptime();
            metrics.total_processes = static_cast<uint32_t>(metrics.processes.size());
            
            uint32_t total_threads = 0;
            for (const auto& process : metrics.processes) {
                total_threads += process.thread_count;
            }
            metrics.total_threads = total_threads;
            
        } catch (const std::exception& e) {
            utils::Logger::error("Failed to collect system metrics: {}", e.what());
        }
        
        return metrics;
    }
    
    std::vector<ProcessInfo> collect_process_info() {
        std::vector<ProcessInfo> processes;
        
#ifdef _WIN32
        HANDLE snapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
        if (snapshot == INVALID_HANDLE_VALUE) {
            utils::Logger::error("Failed to create process snapshot");
            return processes;
        }
        
        PROCESSENTRY32W pe32;
        pe32.dwSize = sizeof(PROCESSENTRY32W);
        
        if (Process32FirstW(snapshot, &pe32)) {
            do {
                ProcessInfo info;
                info.pid = pe32.th32ProcessID;
                
                // Convert wide string to narrow string
                int size = WideCharToMultiByte(CP_UTF8, 0, pe32.szExeFile, -1, nullptr, 0, nullptr, nullptr);
                std::string name(size - 1, 0);
                WideCharToMultiByte(CP_UTF8, 0, pe32.szExeFile, -1, &name[0], size, nullptr, nullptr);
                info.name = name;
                
                info.thread_count = pe32.cntThreads;
                
                // Get additional process information
                HANDLE process_handle = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ, FALSE, pe32.th32ProcessID);
                if (process_handle) {
                    // Get memory information
                    PROCESS_MEMORY_COUNTERS_EX pmc;
                    if (GetProcessMemoryInfo(process_handle, (PROCESS_MEMORY_COUNTERS*)&pmc, sizeof(pmc))) {
                        info.memory_bytes = pmc.WorkingSetSize;
                        info.virtual_memory_bytes = pmc.PrivateUsage;
                    }
                    
                    // Get process times for CPU calculation
                    FILETIME creation_time, exit_time, kernel_time, user_time;
                    if (GetProcessTimes(process_handle, &creation_time, &exit_time, &kernel_time, &user_time)) {
                        // Convert FILETIME to system_clock::time_point
                        ULARGE_INTEGER uli;
                        uli.LowPart = creation_time.dwLowDateTime;
                        uli.HighPart = creation_time.dwHighDateTime;
                        
                        // Convert Windows FILETIME (100ns intervals since 1601) to Unix epoch
                        const uint64_t EPOCH_DIFFERENCE = 11644473600ULL * 10000000ULL;
                        uint64_t time_since_epoch = uli.QuadPart - EPOCH_DIFFERENCE;
                        
                        auto duration = std::chrono::duration<uint64_t, std::ratio<1, 10000000>>(time_since_epoch);
                        info.start_time = std::chrono::system_clock::time_point(
                            std::chrono::duration_cast<std::chrono::system_clock::duration>(duration));
                    }
                    
                    // Get command line (simplified)
                    info.command_line = info.name;
                    info.status = "Running";
                    info.user = "Unknown"; // Would need additional API calls to get actual user
                    
                    CloseHandle(process_handle);
                }
                
                // CPU percentage calculation would require maintaining state between calls
                info.cpu_percent = 0.0; // Placeholder
                
                processes.push_back(info);
                
            } while (Process32NextW(snapshot, &pe32));
        }
        
        CloseHandle(snapshot);
#endif
        
        return processes;
    }
    
    MemoryInfo collect_memory_info() {
        MemoryInfo info = {};
        
#ifdef _WIN32
        MEMORYSTATUSEX mem_status;
        mem_status.dwLength = sizeof(mem_status);
        
        if (GlobalMemoryStatusEx(&mem_status)) {
            info.total_physical = mem_status.ullTotalPhys;
            info.available_physical = mem_status.ullAvailPhys;
            info.used_physical = info.total_physical - info.available_physical;
            info.total_virtual = mem_status.ullTotalVirtual;
            info.available_virtual = mem_status.ullAvailVirtual;
            info.used_virtual = info.total_virtual - info.available_virtual;
            info.total_page_file = mem_status.ullTotalPageFile;
            info.available_page_file = mem_status.ullAvailPageFile;
            info.used_page_file = info.total_page_file - info.available_page_file;
            info.memory_load_percent = static_cast<double>(mem_status.dwMemoryLoad);
        }
#endif
        
        return info;
    }
    
    CpuInfo collect_cpu_info() {
        CpuInfo info = {};
        
#ifdef _WIN32
        SYSTEM_INFO sys_info;
        GetSystemInfo(&sys_info);
        
        info.logical_processor_count = sys_info.dwNumberOfProcessors;
        info.core_count = sys_info.dwNumberOfProcessors; // Simplified
        
        // Get CPU name from registry (simplified)
        info.name = "Windows CPU";
        
        // CPU usage calculation using PDH
        if (cpu_query_ && cpu_counter_) {
            PDH_FMT_COUNTERVALUE counter_value;
            PDH_STATUS status = PdhCollectQueryData(cpu_query_);
            if (status == ERROR_SUCCESS) {
                status = PdhGetFormattedCounterValue(cpu_counter_, PDH_FMT_DOUBLE, nullptr, &counter_value);
                if (status == ERROR_SUCCESS) {
                    info.usage_percent = counter_value.doubleValue;
                }
            }
        }
        
        // Frequency (simplified)
        info.frequency_mhz = 2400; // Placeholder
        info.temperature_celsius = 0.0; // Not easily available on Windows
        
        // Per-core usage would require additional counters
        info.per_core_usage.resize(info.logical_processor_count, info.usage_percent / info.logical_processor_count);
#endif
        
        return info;
    }
    
    std::vector<NetworkInfo> collect_network_info() {
        std::vector<NetworkInfo> interfaces;
        
#ifdef _WIN32
        ULONG buffer_size = 0;
        GetAdaptersInfo(nullptr, &buffer_size);
        
        if (buffer_size > 0) {
            std::vector<char> buffer(buffer_size);
            PIP_ADAPTER_INFO adapter_info = reinterpret_cast<PIP_ADAPTER_INFO>(buffer.data());
            
            if (GetAdaptersInfo(adapter_info, &buffer_size) == ERROR_SUCCESS) {
                PIP_ADAPTER_INFO adapter = adapter_info;
                while (adapter) {
                    NetworkInfo info;
                    info.interface_name = adapter->AdapterName;
                    info.interface_description = adapter->Description;
                    info.is_connected = (adapter->Type != MIB_IF_TYPE_LOOPBACK);
                    info.ip_address = adapter->IpAddressList.IpAddress.String;
                    
                    // Network statistics would require additional API calls
                    info.bytes_sent = 0;
                    info.bytes_received = 0;
                    info.packets_sent = 0;
                    info.packets_received = 0;
                    info.send_rate_bps = 0;
                    info.receive_rate_bps = 0;
                    
                    interfaces.push_back(info);
                    adapter = adapter->Next;
                }
            }
        }
#endif
        
        return interfaces;
    }
    
    std::vector<DiskInfo> collect_disk_info() {
        std::vector<DiskInfo> disks;
        
#ifdef _WIN32
        DWORD drives = GetLogicalDrives();
        char drive_letter = 'A';
        
        for (int i = 0; i < 26; i++) {
            if (drives & (1 << i)) {
                std::string drive_path = std::string(1, drive_letter + i) + ":\\";
                
                UINT drive_type = GetDriveTypeA(drive_path.c_str());
                if (drive_type == DRIVE_FIXED || drive_type == DRIVE_REMOVABLE) {
                    DiskInfo info;
                    info.drive_letter = std::string(1, drive_letter + i);
                    
                    // Get volume information
                    char volume_name[MAX_PATH];
                    char file_system[MAX_PATH];
                    DWORD serial_number, max_component_length, file_system_flags;
                    
                    if (GetVolumeInformationA(drive_path.c_str(), volume_name, MAX_PATH,
                                            &serial_number, &max_component_length,
                                            &file_system_flags, file_system, MAX_PATH)) {
                        info.volume_label = volume_name;
                        info.file_system = file_system;
                    }
                    
                    // Get disk space
                    ULARGE_INTEGER free_bytes, total_bytes;
                    if (GetDiskFreeSpaceExA(drive_path.c_str(), &free_bytes, &total_bytes, nullptr)) {
                        info.total_space = total_bytes.QuadPart;
                        info.free_space = free_bytes.QuadPart;
                        info.used_space = info.total_space - info.free_space;
                        info.usage_percent = (static_cast<double>(info.used_space) / info.total_space) * 100.0;
                    }
                    
                    // Disk I/O statistics would require performance counters
                    info.read_rate_bps = 0;
                    info.write_rate_bps = 0;
                    info.read_iops = 0;
                    info.write_iops = 0;
                    
                    disks.push_back(info);
                }
            }
        }
#endif
        
        return disks;
    }

private:
#ifdef _WIN32
    PDH_HQUERY cpu_query_ = nullptr;
    PDH_HCOUNTER cpu_counter_ = nullptr;
    
    void initialize_pdh() {
        PDH_STATUS status = PdhOpenQueryW(nullptr, 0, &cpu_query_);
        if (status == ERROR_SUCCESS) {
            status = PdhAddEnglishCounterW(cpu_query_, L"\\Processor(_Total)\\% Processor Time", 0, &cpu_counter_);
            if (status == ERROR_SUCCESS) {
                PdhCollectQueryData(cpu_query_); // Initial collection
            }
        }
    }
    
    void cleanup_pdh() {
        if (cpu_query_) {
            PdhCloseQuery(cpu_query_);
        }
    }
#endif
    
    double get_system_uptime() {
#ifdef _WIN32
        return static_cast<double>(GetTickCount64()) / 1000.0;
#else
        return 0.0;
#endif
    }
};

SystemMetricsCollector::SystemMetricsCollector() : pimpl_(std::make_unique<Impl>()) {}

SystemMetricsCollector::~SystemMetricsCollector() = default;

SystemMetrics SystemMetricsCollector::collect() {
    return pimpl_->collect();
}

std::vector<ProcessInfo> SystemMetricsCollector::collect_process_info() {
    return pimpl_->collect_process_info();
}

MemoryInfo SystemMetricsCollector::collect_memory_info() {
    return pimpl_->collect_memory_info();
}

CpuInfo SystemMetricsCollector::collect_cpu_info() {
    return pimpl_->collect_cpu_info();
}

std::vector<NetworkInfo> SystemMetricsCollector::collect_network_info() {
    return pimpl_->collect_network_info();
}

std::vector<DiskInfo> SystemMetricsCollector::collect_disk_info() {
    return pimpl_->collect_disk_info();
}

} // namespace metrics
} // namespace wtop
