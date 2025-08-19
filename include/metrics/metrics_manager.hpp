#pragma once

#include <memory>
#include <thread>
#include <atomic>
#include <queue>
#include <mutex>
#include <condition_variable>
#include "system_metrics.hpp"
#include "utils/config.hpp"

namespace wtop {
namespace metrics {

class MetricsManager {
public:
    explicit MetricsManager(const utils::Config& config);
    ~MetricsManager();
    
    void start();
    void stop();
    
    SystemMetrics get_latest_metrics();
    std::vector<SystemMetrics> get_metrics_history(int seconds = 60);
    
    bool is_running() const { return running_; }

private:
    void collection_loop();
    void cleanup_old_metrics();
    
    const utils::Config& config_;
    std::unique_ptr<SystemMetricsCollector> collector_;
    
    std::atomic<bool> running_{false};
    std::thread collection_thread_;
    
    mutable std::mutex metrics_mutex_;
    std::queue<SystemMetrics> metrics_history_;
    SystemMetrics latest_metrics_;
    
    std::chrono::steady_clock::time_point last_cleanup_;
};

} // namespace metrics
} // namespace wtop
