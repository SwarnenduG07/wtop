#include "utils/logger.hpp"
#include <iostream>
#include <chrono>
#include <iomanip>
#include <mutex>

namespace wtop {
namespace utils {

LogLevel Logger::current_level_ = LogLevel::INFO;
std::mutex log_mutex;

void Logger::initialize(const std::string& level) {
    current_level_ = string_to_level(level);
}

void Logger::log(LogLevel level, const std::string& message) {
    if (level < current_level_) {
        return;
    }
    
    std::lock_guard<std::mutex> lock(log_mutex);
    
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(
        now.time_since_epoch()) % 1000;
    
    std::cerr << std::put_time(std::localtime(&time_t), "%Y-%m-%d %H:%M:%S")
              << "." << std::setfill('0') << std::setw(3) << ms.count()
              << " [" << level_to_string(level) << "] " << message << std::endl;
}

void Logger::debug(const std::string& message) {
    log(LogLevel::DEBUG, message);
}

void Logger::info(const std::string& message) {
    log(LogLevel::INFO, message);
}

void Logger::warn(const std::string& message) {
    log(LogLevel::WARN, message);
}

void Logger::error(const std::string& message) {
    log(LogLevel::ERROR, message);
}

std::string Logger::level_to_string(LogLevel level) {
    switch (level) {
        case LogLevel::DEBUG: return "DEBUG";
        case LogLevel::INFO:  return "INFO ";
        case LogLevel::WARN:  return "WARN ";
        case LogLevel::ERROR: return "ERROR";
        default: return "UNKNOWN";
    }
}

LogLevel Logger::string_to_level(const std::string& level) {
    if (level == "debug") return LogLevel::DEBUG;
    if (level == "info")  return LogLevel::INFO;
    if (level == "warn")  return LogLevel::WARN;
    if (level == "error") return LogLevel::ERROR;
    return LogLevel::INFO;
}

} // namespace utils
} // namespace wtop
