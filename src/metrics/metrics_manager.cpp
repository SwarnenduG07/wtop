#include "metrics/metrics_manager.hpp"
#include "utils/logger.hpp"
#include <algorithm>

namespace wtop {
namespace metrics {

MetricsManager::MetricsManager(const utils::Config& config)
    : config_(config)
    , collector_(std::make_unique<SystemMetricsCollector>())
    , last_cleanup_(std::chrono::steady_clock::now()) {
}

MetricsManager::~MetricsManager() {
    stop();
}

void MetricsManager::start() {
    if (running_) {
        return;
    }
    
    utils::Logger::info("Starting metrics collection");
    running_ = true;
    collection_thread_ = std::thread(&MetricsManager::collection_loop, this);
}

void MetricsManager::stop() {
    if (!running_) {
        return;
    }
    
    utils::Logger::info("Stopping metrics collection");
    running_ = false;
    
    if (collection_thread_.joinable()) {
        collection_thread_.join();
    }
}

SystemMetrics MetricsManager::get_latest_metrics() {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return latest_metrics_;
}

std::vector<SystemMetrics> MetricsManager::get_metrics_history(int seconds) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    
    std::vector<SystemMetrics> history;
    auto cutoff_time = std::chrono::system_clock::now() - std::chrono::seconds(seconds);
    
    // Convert queue to vector for easier processing
    std::queue<SystemMetrics> temp_queue = metrics_history_;
    while (!temp_queue.empty()) {
        const auto& metrics = temp_queue.front();
        if (metrics.timestamp >= cutoff_time) {
            history.push_back(metrics);
        }
        temp_queue.pop();
    }
    
    return history;
}

void MetricsManager::collection_loop() {
    utils::Logger::debug("Metrics collection loop started");
    
    while (running_) {
        try {
            auto start_time = std::chrono::steady_clock::now();
            
            // Collect metrics
            SystemMetrics metrics = collector_->collect();
            
            // Store metrics
            {
                std::lock_guard<std::mutex> lock(metrics_mutex_);
                latest_metrics_ = metrics;
                metrics_history_.push(metrics);
                
                // Limit history size
                while (metrics_history_.size() > static_cast<size_t>(config_.metric_buffer_size)) {
                    metrics_history_.pop();
                }
            }
            
            // Cleanup old metrics periodically
            auto now = std::chrono::steady_clock::now();
            if (now - last_cleanup_ > std::chrono::minutes(1)) {
                cleanup_old_metrics();
                last_cleanup_ = now;
            }
            
            // Calculate sleep time to maintain refresh rate
            auto collection_time = std::chrono::steady_clock::now() - start_time;
            auto target_interval = std::chrono::milliseconds(config_.refresh_rate);
            
            if (collection_time < target_interval) {
                std::this_thread::sleep_for(target_interval - collection_time);
            } else {
                utils::Logger::warn("Metrics collection took {}ms, longer than refresh rate of {}ms",
                                  std::chrono::duration_cast<std::chrono::milliseconds>(collection_time).count(),
                                  config_.refresh_rate);
            }
            
        } catch (const std::exception& e) {
            utils::Logger::error("Error in metrics collection loop: {}", e.what());
            std::this_thread::sleep_for(std::chrono::seconds(1));
        }
    }
    
    utils::Logger::debug("Metrics collection loop stopped");
}

void MetricsManager::cleanup_old_metrics() {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    
    auto cutoff_time = std::chrono::system_clock::now() - 
                      std::chrono::seconds(config_.history_retention_seconds);
    
    // Remove old metrics from history
    std::queue<SystemMetrics> cleaned_history;
    while (!metrics_history_.empty()) {
        const auto& metrics = metrics_history_.front();
        if (metrics.timestamp >= cutoff_time) {
            cleaned_history.push(metrics);
        }
        metrics_history_.pop();
    }
    
    metrics_history_ = std::move(cleaned_history);
    
    utils::Logger::debug("Cleaned up old metrics, {} entries remaining", metrics_history_.size());
}

} // namespace metrics
} // namespace wtop
