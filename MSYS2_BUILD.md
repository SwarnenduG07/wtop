# Building wtop with MSYS2

This guide explains how to build wtop using MSYS2, which provides a Unix-like environment on Windows with the MinGW-w64 toolchain.

## Prerequisites

### 1. Install MSYS2

Download and install MSYS2 from [https://www.msys2.org/](https://www.msys2.org/)

### 2. Update MSYS2

Open MSYS2 terminal and update the package database:

```bash
pacman -Syu
```

Close the terminal when prompted, then reopen and run:

```bash
pacman -Su
```

### 3. Install Build Tools

Install the essential build tools:

```bash
# Install toolchain and build essentials
pacman -S mingw-w64-x86_64-toolchain
pacman -S mingw-w64-x86_64-cmake
pacman -S mingw-w64-x86_64-make
pacman -S mingw-w64-x86_64-pkg-config
pacman -S mingw-w64-x86_64-ninja  # Optional, for faster builds

# Install Git if not already available
pacman -S git
```

### 4. Install Optional Dependencies

```bash
# CLI11 (if available in MSYS2 repos)
pacman -S mingw-w64-x86_64-cli11

# Other useful tools
pacman -S mingw-w64-x86_64-gdb      # For debugging
pacman -S mingw-w64-x86_64-ccache   # For faster rebuilds
```

## Building wtop

### Method 1: Using the MSYS2 Build Script (Recommended)

1. **Open MINGW64 terminal** (not the MSYS2 terminal):
   - Look for "MSYS2 MinGW x64" in your Start Menu
   - Or run `mingw64.exe` from your MSYS2 installation

2. **Navigate to the project directory:**
   ```bash
   cd /d/code/Project/wtop
   ```

3. **Make the build script executable and run it:**
   ```bash
   chmod +x build_msys2.sh
   ./build_msys2.sh
   ```

   For a debug build:
   ```bash
   ./build_msys2.sh Debug
   ```

### Method 2: Manual Build

1. **Open MINGW64 terminal**

2. **Navigate to project and create build directory:**
   ```bash
   cd /d/code/Project/wtop
   mkdir -p build
   cd build
   ```

3. **Configure with CMake:**
   ```bash
   cmake .. -G "Unix Makefiles" -DCMAKE_BUILD_TYPE=Release
   ```

4. **Build:**
   ```bash
   make -j$(nproc)
   ```

### Method 3: Using Ninja (Faster builds)

```bash
cd build
cmake .. -G "Ninja" -DCMAKE_BUILD_TYPE=Release
ninja
```

## Troubleshooting

### Common Issues

1. **"Command not found" errors:**
   - Make sure you're using the MINGW64 terminal, not the MSYS2 terminal
   - Verify tools are installed: `which gcc cmake make`

2. **OpenTelemetry build issues:**
   - The build script automatically disables gRPC exporter for MinGW compatibility
   - If you encounter protobuf issues, the build will fall back to HTTP-only export

3. **Missing dependencies:**
   - The build script will automatically download CLI11 and OpenTelemetry via CMake FetchContent
   - This requires an internet connection during the first build

4. **Linking errors:**
   - Make sure you have the complete mingw-w64 toolchain installed
   - Try a clean build: `rm -rf build && mkdir build`

### Performance Tips

1. **Use ccache for faster rebuilds:**
   ```bash
   pacman -S mingw-w64-x86_64-ccache
   export CC="ccache gcc"
   export CXX="ccache g++"
   ```

2. **Use Ninja instead of Make:**
   ```bash
   pacman -S mingw-w64-x86_64-ninja
   cmake .. -G "Ninja"
   ninja
   ```

3. **Parallel builds:**
   ```bash
   make -j$(nproc)  # Uses all CPU cores
   ```

## Environment Variables

You can set these environment variables to customize the build:

```bash
# Use vcpkg if installed
export VCPKG_ROOT="/path/to/vcpkg"

# Compiler cache
export CCACHE_DIR="$HOME/.ccache"

# Custom CMake options
export CMAKE_ARGS="-DBUILD_TESTING=ON"
```

## Running wtop

After a successful build:

```bash
# From the build directory
./wtop.exe --help

# Or install system-wide
make install
wtop --help
```

## Packaging

To create a distributable package:

```bash
# Create a ZIP package
cpack -G ZIP

# Or create a tarball
cpack -G TGZ
```

## Cross-compilation

MSYS2 also supports cross-compilation for different architectures:

```bash
# For 32-bit Windows
pacman -S mingw-w64-i686-toolchain
# Then use the MINGW32 terminal and build normally
```

## Differences from Visual Studio Build

- **Static linking:** The MSYS2 build uses static linking by default for better portability
- **gRPC disabled:** OpenTelemetry gRPC exporter is disabled due to MinGW compatibility issues
- **Different runtime:** Uses MinGW runtime instead of MSVC runtime
- **Better POSIX compatibility:** More Unix-like behavior in some system calls

## Benefits of MSYS2 Build

- **No Visual Studio required:** Lighter development environment
- **Better package management:** Easy dependency installation with pacman
- **Cross-platform familiarity:** Similar to Linux development workflow
- **Static binaries:** Easier distribution without runtime dependencies
- **Open source toolchain:** Completely free and open source build environment
