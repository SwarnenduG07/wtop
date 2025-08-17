#!/bin/bash

set -e  # Exit on any error

echo "Building wtop - Windows System Monitor (MSYS2)"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default build type
BUILD_TYPE=${1:-Release}

echo -e "${BLUE}Build type: $BUILD_TYPE${NC}"

# Check if we're in MSYS2 environment
if [[ -z "$MSYSTEM" ]]; then
    echo -e "${RED}Error: Not running in MSYS2 environment${NC}"
    echo "Please run this script from MSYS2 terminal (MINGW64 or UCRT64)"
    exit 1
fi

echo -e "${GREEN}MSYS2 environment detected: $MSYSTEM${NC}"

# Check for required tools
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: $1 not found${NC}"
        echo "Install with: pacman -S $2"
        return 1
    else
        echo -e "${GREEN}✓ $1 found${NC}"
        return 0
    fi
}

echo "Checking required tools..."

# Check for essential build tools
MISSING_TOOLS=0

if ! check_tool "gcc" "mingw-w64-x86_64-gcc"; then
    MISSING_TOOLS=1
fi

if ! check_tool "g++" "mingw-w64-x86_64-gcc"; then
    MISSING_TOOLS=1
fi

if ! check_tool "cmake" "mingw-w64-x86_64-cmake"; then
    MISSING_TOOLS=1
fi

if ! check_tool "make" "mingw-w64-x86_64-make"; then
    MISSING_TOOLS=1
fi

if ! check_tool "pkg-config" "mingw-w64-x86_64-pkg-config"; then
    MISSING_TOOLS=1
fi

if [[ $MISSING_TOOLS -eq 1 ]]; then
    echo -e "${RED}Missing required tools. Install them with:${NC}"
    echo "pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-cmake mingw-w64-x86_64-make mingw-w64-x86_64-pkg-config"
    exit 1
fi

# Check for optional dependencies
echo "Checking optional dependencies..."

if command -v "vcpkg" &> /dev/null; then
    echo -e "${GREEN}✓ vcpkg found${NC}"
    USE_VCPKG=1
else
    echo -e "${YELLOW}! vcpkg not found, will try to use system packages${NC}"
    USE_VCPKG=0
fi

# Check for CLI11
if pkg-config --exists cli11 2>/dev/null; then
    echo -e "${GREEN}✓ CLI11 found via pkg-config${NC}"
elif [[ -d "/mingw64/include/CLI" ]]; then
    echo -e "${GREEN}✓ CLI11 found in system${NC}"
else
    echo -e "${YELLOW}! CLI11 not found, will download via FetchContent${NC}"
fi

# Check for OpenTelemetry (this is usually not available in MSYS2 packages)
if pkg-config --exists opentelemetry-cpp 2>/dev/null; then
    echo -e "${GREEN}✓ OpenTelemetry C++ found${NC}"
else
    echo -e "${YELLOW}! OpenTelemetry C++ not found, will download via FetchContent${NC}"
fi

# Create build directory
echo "Creating build directory..."
mkdir -p build
cd build

# Configure CMake
echo -e "${BLUE}Configuring CMake...${NC}"

CMAKE_ARGS=(
    "-DCMAKE_BUILD_TYPE=$BUILD_TYPE"
    "-DCMAKE_CXX_STANDARD=17"
    "-G" "Unix Makefiles"
)

# Add vcpkg toolchain if available
if [[ $USE_VCPKG -eq 1 && -n "$VCPKG_ROOT" ]]; then
    echo -e "${GREEN}Using vcpkg toolchain${NC}"
    CMAKE_ARGS+=("-DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake")
fi

# Add Windows-specific definitions
CMAKE_ARGS+=(
    "-DCMAKE_SYSTEM_NAME=Windows"
    "-DWIN32=1"
)

echo "CMake command: cmake .. ${CMAKE_ARGS[*]}"
cmake .. "${CMAKE_ARGS[@]}"

if [[ $? -ne 0 ]]; then
    echo -e "${RED}CMake configuration failed${NC}"
    exit 1
fi

# Build
echo -e "${BLUE}Building wtop...${NC}"
make -j$(nproc)

if [[ $? -ne 0 ]]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${BLUE}Executable location: $(pwd)/wtop.exe${NC}"

# Test if the executable was created
if [[ -f "wtop.exe" ]]; then
    echo -e "${GREEN}✓ wtop.exe created successfully${NC}"
    
    # Show file info
    ls -la wtop.exe
    
    echo ""
    echo -e "${YELLOW}Usage examples:${NC}"
    echo "  ./wtop.exe --help          Show help"
    echo "  ./wtop.exe                 Start with default settings"
    echo "  ./wtop.exe --output json   JSON output mode"
    echo "  ./wtop.exe --refresh 500   Custom refresh rate"
else
    echo -e "${RED}Error: wtop.exe not found after build${NC}"
    exit 1
fi

cd ..
