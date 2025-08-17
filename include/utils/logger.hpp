#pragma once

#include <string>
#include <memory>
#include <sstream>

namespace wtop {
namespace utils {

enum class LogLevel {
    DEBUG = 0,
    INFO = 1,
    WARN = 2,
    ERROR = 3
};

class Logger {
public:
    static void initialize(const std::string& level);
    static void log(LogLevel level, const std::string& message);
    static void debug(const std::string& message);
    static void info(const std::string& message);
    static void warn(const std::string& message);
    static void error(const std::string& message);
    
    template<typename... Args>
    static void debug(const std::string& format, Args&&... args) {
        log(LogLevel::DEBUG, format_string(format, std::forward<Args>(args)...));
    }
    
    template<typename... Args>
    static void info(const std::string& format, Args&&... args) {
        log(LogLevel::INFO, format_string(format, std::forward<Args>(args)...));
    }
    
    template<typename... Args>
    static void warn(const std::string& format, Args&&... args) {
        log(LogLevel::WARN, format_string(format, std::forward<Args>(args)...));
    }
    
    template<typename... Args>
    static void error(const std::string& format, Args&&... args) {
        log(LogLevel::ERROR, format_string(format, std::forward<Args>(args)...));
    }

private:
    static LogLevel current_level_;
    static std::string level_to_string(LogLevel level);
    static LogLevel string_to_level(const std::string& level);
    
    template<typename... Args>
    static std::string format_string(const std::string& format, Args&&... args) {
        std::ostringstream oss;
        format_impl(oss, format, std::forward<Args>(args)...);
        return oss.str();
    }
    
    template<typename T, typename... Args>
    static void format_impl(std::ostringstream& oss, const std::string& format, T&& value, Args&&... args) {
        size_t pos = format.find("{}");
        if (pos != std::string::npos) {
            oss << format.substr(0, pos) << std::forward<T>(value);
            format_impl(oss, format.substr(pos + 2), std::forward<Args>(args)...);
        } else {
            oss << format;
        }
    }
    
    static void format_impl(std::ostringstream& oss, const std::string& format) {
        oss << format;
    }
};

} // namespace utils
} // namespace wtop
