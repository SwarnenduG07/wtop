# wtop - Windows System Monitor

A modern, interactive system monitoring tool for Windows, similar to htop but designed specifically for Windows systems. Built with OpenTelemetry for metrics collection and CLI11 for command-line interface.

## Features

- **Real-time System Monitoring**: CPU, memory, disk, and network usage
- **Interactive Process Management**: Sort, filter, and monitor processes
- **Multiple Display Modes**: Overview, processes, memory, CPU, network, and disk views
- **OpenTelemetry Integration**: Export metrics to OTLP-compatible backends
- **Flexible Output Formats**: Interactive terminal UI, JSON, and CSV output
- **Windows-Optimized**: Native Windows API integration for accurate metrics

## Screenshots

```
wtop - Windows System Monitor                                    14:30:25 up 2:15
Tasks: 156 total, 1024 threads                                   Load: 23.5%
--------------------------------------------------------------------------------
CPU: Intel Core i7-10700K
Usage: 23.5% (8 cores)

Memory: 8.2 GB / 16.0 GB (51.2%)

Top Processes by CPU:
PID     Name                 CPU%    Memory
1234    chrome.exe           15.2    512.3 MB
5678    code.exe             8.7     256.1 MB
9012    wtop.exe             2.1     32.5 MB
```

## Installation

### Prerequisites

- Windows 7 or later
- Visual Studio 2019 or later (for building from source)
- CMake 3.16 or later
- vcpkg (recommended for dependencies)

### Using vcpkg (Recommended)

```bash
# Install dependencies
vcpkg install cli11 opentelemetry-cpp

# Build
mkdir build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=[vcpkg root]/scripts/buildsystems/vcpkg.cmake
cmake --build . --config Release
```

### Manual Build

```bash
# Clone repository
git clone https://github.com/your-org/wtop.git
cd wtop

# Create build directory
mkdir build && cd build

# Configure
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
cmake --build . --config Release

# Install (optional)
cmake --install .
```

## Usage

### Basic Usage

```bash
# Start wtop with default settings
wtop

# Custom refresh rate (500ms)
wtop --refresh 500

# JSON output mode
wtop --output json

# CSV output mode
wtop --output csv

# Disable telemetry
wtop --no-telemetry

# Show only specific metrics
wtop --no-network --no-disk
```

### Interactive Controls

| Key | Action |
|-----|--------|
| `1-6` | Switch between views (Overview, Processes, Memory, CPU, Network, Disk) |
| `h`, `?` | Show/hide help |
| `q`, `ESC` | Quit |
| `p` | Sort processes by PID |
| `n` | Sort processes by Name |
| `c` | Sort processes by CPU usage |
| `m` | Sort processes by Memory usage |
| `t` | Sort processes by Thread count |
| `r` | Reverse sort order |
| `↑`/`↓` | Scroll process list |
| `/` | Filter processes |

### Command Line Options

```
wtop - Windows System Monitor

USAGE:
  wtop [OPTIONS]

OPTIONS:
  -h,--help                   Print this help message and exit
  -r,--refresh INT:INT in [100 - 10000]
                              Refresh rate in milliseconds (default: 1000)
  --no-telemetry              Disable OpenTelemetry metrics collection
  -l,--log-level ENUM:value in {debug->0,info->1,warn->2,error->3} OR {0,1,2,3}
                              Log level (debug, info, warn, error)
  -o,--output ENUM:value in {interactive->0,json->1,csv->2} OR {0,1,2}
                              Output format (interactive, json, csv)
  --no-processes              Hide process information
  --no-memory                 Hide memory information
  --no-cpu                    Hide CPU information
  --no-network                Hide network information
  --no-disk                   Hide disk information
```

## Configuration

### OpenTelemetry Configuration

wtop supports exporting metrics to OpenTelemetry-compatible backends:

```bash
# Set OTLP endpoint
set OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# Set service name
set OTEL_SERVICE_NAME=wtop

# Set resource attributes
set OTEL_RESOURCE_ATTRIBUTES=service.version=1.0.0,deployment.environment=production
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `WTOP_REFRESH_RATE` | Refresh rate in milliseconds | 1000 |
| `WTOP_LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `WTOP_TELEMETRY_ENDPOINT` | OpenTelemetry endpoint | http://localhost:4317 |
| `WTOP_NO_COLOR` | Disable colored output | false |

## Metrics

wtop collects and can export the following metrics via OpenTelemetry:

### System Metrics
- `system.cpu.usage` - CPU usage percentage
- `system.memory.usage` - Memory usage in bytes
- `system.memory.available` - Available memory in bytes
- `system.network.bytes` - Network bytes transferred
- `system.disk.io` - Disk I/O operations
- `system.disk.free` - Free disk space in bytes

### Process Metrics
- `process.cpu.usage` - Per-process CPU usage
- `process.memory.usage` - Per-process memory usage
- `system.process.count` - Total number of processes

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI11         │    │  OpenTelemetry   │    │  Windows APIs   │
│  (Arguments)    │    │   (Metrics)      │    │   (Data)        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                        wtop Core                                │
├─────────────────┬──────────────────┬─────────────────────────────┤
│   UI Display    │  Metrics Manager │    Telemetry Manager        │
│   - Terminal    │  - Collection    │    - OTLP Export            │
│   - Formatting  │  - Aggregation   │    - Instrumentation        │
│   - Interaction │  - History       │    - Tracing                │
└─────────────────┴──────────────────┴─────────────────────────────┘
```

## Development

### Project Structure

```
wtop/
├── include/
│   ├── metrics/
│   │   ├── system_metrics.hpp
│   │   └── metrics_manager.hpp
│   ├── telemetry/
│   │   └── telemetry_manager.hpp
│   ├── ui/
│   │   └── display.hpp
│   └── utils/
│       ├── config.hpp
│       └── logger.hpp
├── src/
│   ├── metrics/
│   ├── telemetry/
│   ├── ui/
│   └── utils/
├── main.cpp
├── CMakeLists.txt
└── README.md
```

### Building for Development

```bash
# Debug build
cmake .. -DCMAKE_BUILD_TYPE=Debug -DBUILD_TESTS=ON

# Build with tests
cmake --build . --config Debug
ctest
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [CLI11](https://github.com/CLIUtils/CLI11) - Command line parser
- [OpenTelemetry C++](https://github.com/open-telemetry/opentelemetry-cpp) - Observability framework
- [htop](https://htop.dev/) - Inspiration for the interface design

## Support

- GitHub Issues: [Report bugs and request features](https://github.com/your-org/wtop/issues)
- Documentation: [Wiki](https://github.com/your-org/wtop/wiki)
- Discussions: [GitHub Discussions](https://github.com/your-org/wtop/discussions)
