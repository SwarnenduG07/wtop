# wtop - Cross-Platform System Monitor

A lightweight, interactive system monitoring tool similar to htop, built in Go. Provides real-time CPU, memory, disk, and network usage with process monitoring.

## Features

- **Real-time System Monitoring**: CPU, memory, disk, and network usage
- **Interactive Process List**: View top processes sorted by CPU usage
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **Lightweight**: Single executable, no dependencies required
- **Clean Interface**: Terminal-based UI with auto-refresh

## Screenshots

```
wtop - System Monitor                                           01:16:48
================================================================================
CPU: 15.2% (8 cores)
Memory: 8.2 GB / 16.0 GB (51.2%)
Disk (C:\): 245.3 GB / 500.0 GB (49.1%)
Network: ↑ 125.4 MB ↓ 1024.7 MB

Top Processes by CPU:
PID      Name                      CPU%     Memory    
------------------------------------------------------------------------
1234     chrome.exe                15.2     512.3      MB
5678     code.exe                  8.7      256.1      MB
9012     wtop.exe                  2.1      32.5       MB
4567     discord.exe               1.8      145.2      MB
```

## Installation

### Download Pre-built Binaries

Download the appropriate executable for your platform:
- **Windows**: `wtop-windows.exe`
- **Linux**: `wtop-linux`
- **macOS**: `wtop-macos`

### Build from Source

#### Prerequisites
- Go 1.18 or later

#### Build Commands

```bash
# Clone repository
git clone https://github.com/SwarnenduG07/wtop
cd wtop

# Install dependencies
go mod tidy

# Build for current platform
go build -o wtop main.go

# Build for specific platforms
GOOS=windows GOARCH=amd64 go build -o wtop-windows.exe main.go
GOOS=linux GOARCH=amd64 go build -o wtop-linux main.go
GOOS=darwin GOARCH=amd64 go build -o wtop-macos main.go
```

## Usage

### Running wtop

#### Windows
```cmd
# Command Prompt or PowerShell
wtop-windows.exe

# Or double-click the executable
```

#### Linux/macOS
```bash
# Make executable (first time only)
chmod +x wtop-linux

# Run
./wtop-linux
```

### Controls

- **Ctrl+C** or **Ctrl+D**: Exit wtop
- Auto-refreshes every 3 seconds

## System Requirements

- **Windows**: Windows 7 or later
- **Linux**: Any modern Linux distribution
- **macOS**: macOS 10.12 or later
- **Architecture**: 64-bit (amd64)

## What wtop Shows

### System Metrics
- **CPU Usage**: Overall CPU percentage and core count
- **Memory**: Used/Total memory in GB with percentage
- **Disk**: Used/Total disk space with percentage
- **Network**: Total bytes sent (↑) and received (↓)

### Process Information
- **PID**: Process ID
- **Name**: Process name (truncated if too long)
- **CPU%**: Current CPU usage percentage
- **Memory**: Memory usage in MB

## Dependencies

wtop uses minimal dependencies:
- `github.com/shirou/gopsutil/v3` - Cross-platform system metrics

## Project Structure

```
wtop/
├── main.go                 # Main source code
├── go.mod                  # Go module file
├── go.sum                  # Dependency checksums
├── build.sh               # Cross-platform build script
├── build-windows.bat      # Windows build script
├── wtop-windows.exe       # Windows executable
├── wtop-linux            # Linux executable
├── wtop-macos            # macOS executable
└── README.md             # This file
```

## Development

### Local Development

```bash
# Run directly with Go
go run main.go

# Build and run
go build -o wtop main.go
./wtop
```

### Code Structure

- **System Metrics**: CPU, memory, disk, network collection
- **Process Management**: Process enumeration and sorting
- **Display**: Terminal UI with cross-platform screen clearing
- **Error Handling**: Graceful handling of unavailable metrics

## Performance

- **Memory Usage**: ~5-10 MB
- **CPU Impact**: Minimal (<1% on most systems)
- **Refresh Rate**: 3 seconds (configurable in code)
- **Process Limit**: Shows top 15 processes by CPU usage

## Troubleshooting

### Common Issues

1. **Permission Denied (Linux/macOS)**
   ```bash
   chmod +x wtop-linux
   ```

2. **"Cannot execute binary file"**
   - Ensure you're using the correct binary for your architecture
   - Download the appropriate version for your OS

3. **Metrics showing "N/A"**
   - Some metrics may be unavailable on certain systems
   - This is normal and doesn't affect other functionality

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on multiple platforms
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [gopsutil](https://github.com/shirou/gopsutil) - Cross-platform system metrics library
- [htop](https://htop.dev/) - Inspiration for the interface design

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/wtop/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/wtop/discussions)

---

**wtop** - Simple, fast, cross-platform system monitoring in your terminal.
