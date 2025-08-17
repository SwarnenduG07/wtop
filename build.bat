@echo off
setlocal enabledelayedexpansion

echo Building wtop - Windows System Monitor
echo =====================================

:: Check if we're in a Visual Studio environment
if not defined VCINSTALLDIR (
    echo.
    echo Error: Visual Studio environment not detected.
    echo Please run this script from a Visual Studio Developer Command Prompt.
    echo.
    echo To fix this:
    echo 1. Open "Developer Command Prompt for VS 2019/2022" from Start Menu
    echo 2. Navigate to this directory: cd /d "%~dp0"
    echo 3. Run: build.bat
    echo.
    echo Alternatively, you can run:
    echo   "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"
    echo   (adjust path for your VS version)
    echo.
    pause
    exit /b 1
)

:: Check if CMake is available
cmake --version >nul 2>&1
if errorlevel 1 (
    echo Error: CMake not found. Please install CMake and add it to PATH.
    exit /b 1
)

:: Create build directory
if not exist build mkdir build
cd build

:: Configure build type
set BUILD_TYPE=Release
if "%1"=="debug" set BUILD_TYPE=Debug
if "%1"=="Debug" set BUILD_TYPE=Debug

echo Configuring build type: %BUILD_TYPE%

:: Check for vcpkg
if defined VCPKG_ROOT (
    echo Using vcpkg from: %VCPKG_ROOT%
    cmake .. -G "Visual Studio 17 2022" -A x64 -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake
) else (
    echo vcpkg not found, using system packages
    cmake .. -G "Visual Studio 17 2022" -A x64
)

if errorlevel 1 (
    echo Error: CMake configuration failed
    exit /b 1
)

:: Build the project
echo Building wtop...
cmake --build . --config %BUILD_TYPE%

if errorlevel 1 (
    echo Error: Build failed
    exit /b 1
)

echo.
echo Build completed successfully!
echo Executable location: build\%BUILD_TYPE%\wtop.exe
echo.
echo Usage:
echo   wtop --help          Show help
echo   wtop                 Start with default settings
echo   wtop --output json   JSON output mode
echo   wtop --refresh 500   Custom refresh rate

cd ..
