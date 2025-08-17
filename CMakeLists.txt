cmake_minimum_required(VERSION 3.16)

# Detect build environment
if(WIN32)
    if(MINGW OR MSYS)
        message(STATUS "Building with MinGW/MSYS2")
        set(USING_MINGW TRUE)
    else()
        message(STATUS "Building with MSVC")
        set(USING_MSVC TRUE)
        # Set the generator to Visual Studio if on Windows and not already specified
        if(NOT CMAKE_GENERATOR)
            set(CMAKE_GENERATOR "Visual Studio 17 2022" CACHE STRING "CMake generator" FORCE)
        endif()
    endif()
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
if(USING_MINGW)
    # MinGW/MSYS2 specific options
    add_compile_options(-Wall -Wextra -Wpedantic)
    add_compile_definitions(_WIN32_WINNT=0x0601) # Windows 7+
    
    # Enable static linking for better portability
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -static-libgcc -static-libstdc++")
    endif()
    
elseif(USING_MSVC)
    # MSVC specific options
    add_compile_options(/W4)
    add_definitions(-D_WIN32_WINNT=0x0601) # Windows 7+
    
else()
    # Generic options for other compilers
    add_compile_options(-Wall -Wextra -Wpedantic)
endif()

# Find required packages
find_package(Threads REQUIRED)

# Find CLI11
find_package(CLI11 CONFIG QUIET)
if(NOT CLI11_FOUND)
    # Try pkg-config (common in MSYS2)
    find_package(PkgConfig QUIET)
    if(PkgConfig_FOUND)
        pkg_check_modules(CLI11 QUIET cli11)
    endif()
    
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
endif()

# Find OpenTelemetry (usually needs to be built from source)
find_package(opentelemetry-cpp CONFIG QUIET)
if(NOT opentelemetry-cpp_FOUND)
    message(STATUS "OpenTelemetry C++ not found, using FetchContent")
    include(FetchContent)
    
    # Set OpenTelemetry options before fetching
    set(WITH_OTLP_GRPC ON CACHE BOOL "Enable OTLP gRPC exporter")
    set(WITH_OTLP_HTTP ON CACHE BOOL "Enable OTLP HTTP exporter")
    set(BUILD_TESTING OFF CACHE BOOL "Disable testing")
    set(WITH_EXAMPLES OFF CACHE BOOL "Disable examples")
    set(WITH_BENCHMARK OFF CACHE BOOL "Disable benchmarks")
    
    # For MinGW, we might need to disable some features
    if(USING_MINGW)
        set(WITH_OTLP_GRPC OFF CACHE BOOL "Disable OTLP gRPC exporter for MinGW")
        message(STATUS "Disabling OTLP gRPC exporter for MinGW compatibility")
    endif()
    
    FetchContent_Declare(
        opentelemetry-cpp
        GIT_REPOSITORY https://github.com/open-telemetry/opentelemetry-cpp.git
        GIT_TAG v1.11.0
    )
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
target_link_libraries(wtop PRIVATE Threads::Threads)

# Link CLI11
if(TARGET CLI11::CLI11)
    target_link_libraries(wtop PRIVATE CLI11::CLI11)
elseif(CLI11_FOUND)
    target_include_directories(wtop PRIVATE ${CLI11_INCLUDE_DIRS})
    target_link_libraries(wtop PRIVATE ${CLI11_LIBRARIES})
endif()

# OpenTelemetry libraries - conditional linking based on availability
if(TARGET opentelemetry-cpp::api)
    target_link_libraries(wtop PRIVATE
        opentelemetry-cpp::api
        opentelemetry-cpp::sdk
        opentelemetry-cpp::ext
        opentelemetry-cpp::metrics
        opentelemetry-cpp::trace
    )
    
    # Only link OTLP exporters if available
    if(TARGET opentelemetry-cpp::otlp_http_exporter)
        target_link_libraries(wtop PRIVATE opentelemetry-cpp::otlp_http_exporter)
    endif()
    
    if(TARGET opentelemetry-cpp::otlp_grpc_exporter AND NOT USING_MINGW)
        target_link_libraries(wtop PRIVATE opentelemetry-cpp::otlp_grpc_exporter)
    endif()
    
elseif(TARGET opentelemetry_api)
    # Fallback for different target names
    target_link_libraries(wtop PRIVATE
        opentelemetry_api
        opentelemetry_sdk
        opentelemetry_metrics
        opentelemetry_trace
    )
    
    if(TARGET opentelemetry_exporter_otlp_http)
        target_link_libraries(wtop PRIVATE opentelemetry_exporter_otlp_http)
    endif()
    
    if(TARGET opentelemetry_exporter_otlp_grpc AND NOT USING_MINGW)
        target_link_libraries(wtop PRIVATE opentelemetry_exporter_otlp_grpc)
    endif()
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
