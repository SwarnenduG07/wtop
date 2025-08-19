#include "utils/logger.hpp"
#include <iostream>

namespace wtop {
namespace utils {

void Logger::initialize(const std::string& level) {
    std::cout << "Logger initialized with level: " << level << std::endl;
}

void Logger::log(LogLevel level, const std::string& message) {
    std::cout << "[LOG] " << message << std::endl;
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

} // namespace utils
} // namespace wtop
