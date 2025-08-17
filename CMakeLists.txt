cmake_minimum_required(VERSION 3.16)

# Set the generator to Visual Studio if on Windows and not already specified
if(WIN32 AND NOT CMAKE_GENERATOR)
    set(CMAKE_GENERATOR "Visual Studio 17 2022" CACHE STRING "CMake generator" FORCE)
endif()

project(wtop VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Set build type if not specified
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# Compiler-specific options
if(MSVC)
    add_compile_options(/W4 /WX)
    add_definitions(-D_WIN32_WINNT=0x0601) # Windows 7+
else()
    add_compile_options(-Wall -Wextra -Wpedantic -Werror)
endif()

# Find required packages
find_package(Threads REQUIRED)

# Find CLI11
find_package(CLI11 CONFIG REQUIRED)
if(NOT CLI11_FOUND)
    message(STATUS "CLI11 not found, using FetchContent")
    include(FetchContent)
    FetchContent_Declare(
        CLI11
        GIT_REPOSITORY https://github.com/CLIUtils/CLI11.git
        GIT_TAG v2.3.2
    )
    FetchContent_MakeAvailable(CLI11)
endif()

# Find OpenTelemetry
find_package(opentelemetry-cpp CONFIG REQUIRED)
if(NOT opentelemetry-cpp_FOUND)
    message(STATUS "OpenTelemetry C++ not found, using FetchContent")
    include(FetchContent)
    FetchContent_Declare(
        opentelemetry-cpp
        GIT_REPOSITORY https://github.com/open-telemetry/opentelemetry-cpp.git
        GIT_TAG v1.11.0
    )
    set(WITH_OTLP_GRPC ON CACHE BOOL "Enable OTLP gRPC exporter")
    set(WITH_OTLP_HTTP ON CACHE BOOL "Enable OTLP HTTP exporter")
    set(BUILD_TESTING OFF CACHE BOOL "Disable testing")
    set(WITH_EXAMPLES OFF CACHE BOOL "Disable examples")
    FetchContent_MakeAvailable(opentelemetry-cpp)
endif()

# Include directories
include_directories(${CMAKE_SOURCE_DIR}/include)

# Source files
set(WTOP_SOURCES
    main.cpp
    src/utils/logger.cpp
    src/metrics/system_metrics.cpp
    src/metrics/metrics_manager.cpp
    src/telemetry/telemetry_manager.cpp
    src/ui/display.cpp
)

# Create executable
add_executable(wtop ${WTOP_SOURCES})

# Link libraries
target_link_libraries(wtop
    PRIVATE
    Threads::Threads
    CLI11::CLI11
)

# OpenTelemetry libraries
if(TARGET opentelemetry-cpp::api)
    target_link_libraries(wtop PRIVATE
        opentelemetry-cpp::api
        opentelemetry-cpp::sdk
        opentelemetry-cpp::ext
        opentelemetry-cpp::otlp_grpc_exporter
        opentelemetry-cpp::otlp_http_exporter
        opentelemetry-cpp::metrics
        opentelemetry-cpp::trace
    )
else()
    # Fallback for manual installation
    target_link_libraries(wtop PRIVATE
        opentelemetry_api
        opentelemetry_sdk
        opentelemetry_exporter_otlp_grpc
        opentelemetry_exporter_otlp_http
        opentelemetry_metrics
        opentelemetry_trace
    )
endif()

# Windows-specific libraries
if(WIN32)
    target_link_libraries(wtop PRIVATE
        pdh
        iphlpapi
        psapi
        ws2_32
    )
endif()

# Compiler definitions
target_compile_definitions(wtop PRIVATE
    $<$<CONFIG:Debug>:DEBUG>
    $<$<CONFIG:Release>:NDEBUG>
)

# Installation
install(TARGETS wtop
    RUNTIME DESTINATION bin
)

# Package configuration
set(CPACK_PACKAGE_NAME "wtop")
set(CPACK_PACKAGE_VERSION "${PROJECT_VERSION}")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "Windows System Monitor CLI Tool")
set(CPACK_PACKAGE_VENDOR "wtop")
set(CPACK_PACKAGE_CONTACT "support@wtop.dev")

if(WIN32)
    set(CPACK_GENERATOR "ZIP;NSIS")
    set(CPACK_NSIS_DISPLAY_NAME "wtop - Windows System Monitor")
    set(CPACK_NSIS_PACKAGE_NAME "wtop")
    set(CPACK_NSIS_CONTACT "support@wtop.dev")
    set(CPACK_NSIS_HELP_LINK "https://github.com/wtop/wtop")
else()
    set(CPACK_GENERATOR "TGZ;DEB;RPM")
    set(CPACK_DEBIAN_PACKAGE_MAINTAINER "wtop <support@wtop.dev>")
    set(CPACK_RPM_PACKAGE_LICENSE "MIT")
endif()

include(CPack)

# Testing (optional)
option(BUILD_TESTS "Build tests" OFF)
if(BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()

# Documentation
option(BUILD_DOCS "Build documentation" OFF)
if(BUILD_DOCS)
    find_package(Doxygen)
    if(DOXYGEN_FOUND)
        configure_file(${CMAKE_CURRENT_SOURCE_DIR}/Doxyfile.in
                      ${CMAKE_CURRENT_BINARY_DIR}/Doxyfile @ONLY)
        add_custom_target(docs
            ${DOXYGEN_EXECUTABLE} ${CMAKE_CURRENT_BINARY_DIR}/Doxyfile
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            COMMENT "Generating API documentation with Doxygen" VERBATIM
        )
    endif()
endif()

# Print build information
message(STATUS "")
message(STATUS "wtop Build Configuration:")
message(STATUS "  Version: ${PROJECT_VERSION}")
message(STATUS "  Build Type: ${CMAKE_BUILD_TYPE}")
message(STATUS "  C++ Standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "  Install Prefix: ${CMAKE_INSTALL_PREFIX}")
message(STATUS "  Build Tests: ${BUILD_TESTS}")
message(STATUS "  Build Docs: ${BUILD_DOCS}")
message(STATUS "")
