#include "metrics/metrics_manager.hpp"
#include <iostream>

namespace wtop {
namespace metrics {

MetricsManager::MetricsManager(const utils::Config& config) 
    : config_(config), collector_(std::make_unique<SystemMetricsCollector>()), running_(false) {
    std::cout << "MetricsManager created" << std::endl;
}

MetricsManager::~MetricsManager() {
    stop();
    std::cout << "MetricsManager destroyed" << std::endl;
}

void MetricsManager::start() {
    if (running_.load()) {
        return;
    }
    
    running_.store(true);
    collection_thread_ = std::thread(&MetricsManager::collection_loop, this);
    std::cout << "MetricsManager started" << std::endl;
}

void MetricsManager::stop() {
    if (!running_.load()) {
        return;
    }
    
    running_.store(false);
    
    if (collection_thread_.joinable()) {
        collection_thread_.join();
    }
    
    std::cout << "MetricsManager stopped" << std::endl;
}

SystemMetrics MetricsManager::get_latest_metrics() {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return latest_metrics_;
}

std::vector<SystemMetrics> MetricsManager::get_metrics_history(int seconds) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    std::vector<SystemMetrics> result;
    
    auto queue_copy = metrics_history_;
    while (!queue_copy.empty()) {
        result.push_back(queue_copy.front());
        queue_copy.pop();
    }
    
    return result;
}

void MetricsManager::collection_loop() {
    while (running_.load()) {
        auto metrics = collector_->collect();
        
        {
            std::lock_guard<std::mutex> lock(metrics_mutex_);
            latest_metrics_ = metrics;
            metrics_history_.push(metrics);
            
            // Keep only last 60 entries
            while (metrics_history_.size() > 60) {
                metrics_history_.pop();
            }
        }
        
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
    }
}

void MetricsManager::cleanup_old_metrics() {
    // Implementation for cleanup if needed
}

} // namespace metrics
} // namespace wtop
