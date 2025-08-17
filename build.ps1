param(
    [string]$BuildType = "Release",
    [switch]$UseVcpkg = $false,
    [switch]$Help = $false
)

if ($Help) {
    Write-Host "wtop Build Script"
    Write-Host "Usage: .\build.ps1 [-BuildType <Debug|Release>] [-UseVcpkg] [-Help]"
    Write-Host ""
    Write-Host "Parameters:"
    Write-Host "  -BuildType    Build configuration (Debug or Release, default: Release)"
    Write-Host "  -UseVcpkg     Use vcpkg for dependencies"
    Write-Host "  -Help         Show this help message"
    exit 0
}

Write-Host "Building wtop - Windows System Monitor" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

# Check if we're running in PowerShell
if ($PSVersionTable.PSVersion.Major -lt 5) {
    Write-Error "PowerShell 5.0 or later is required"
    exit 1
}

# Function to find Visual Studio
function Find-VisualStudio {
    $vsPaths = @(
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Community\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Professional\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles}\Microsoft Visual Studio\2022\Enterprise\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\Community\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\Professional\Common7\Tools\VsDevCmd.bat",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2019\Enterprise\Common7\Tools\VsDevCmd.bat"
    )
    
    foreach ($path in $vsPaths) {
        if (Test-Path $path) {
            return $path
        }
    }
    return $null
}

# Check for Visual Studio
$vsPath = Find-VisualStudio
if (-not $vsPath) {
    Write-Error "Visual Studio 2019 or 2022 not found. Please install Visual Studio with C++ development tools."
    Write-Host "Download from: https://visualstudio.microsoft.com/downloads/" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found Visual Studio at: $vsPath" -ForegroundColor Cyan

# Check for CMake
try {
    $cmakeVersion = & cmake --version 2>$null
    Write-Host "CMake found: $($cmakeVersion[0])" -ForegroundColor Green
} catch {
    Write-Error "CMake not found. Please install CMake and add it to PATH."
    Write-Host "Download from: https://cmake.org/download/" -ForegroundColor Yellow
    exit 1
}

# Create build directory
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Name "build" | Out-Null
}

Set-Location "build"

try {
    Write-Host "Configuring build type: $BuildType" -ForegroundColor Cyan
    
    # Prepare CMake arguments
    $cmakeArgs = @(
        "..",
        "-G", "Visual Studio 17 2022",
        "-A", "x64"
    )
    
    if ($UseVcpkg -and $env:VCPKG_ROOT) {
        Write-Host "Using vcpkg from: $env:VCPKG_ROOT" -ForegroundColor Cyan
        $cmakeArgs += "-DCMAKE_TOOLCHAIN_FILE=$env:VCPKG_ROOT\scripts\buildsystems\vcpkg.cmake"
    } elseif ($UseVcpkg) {
        Write-Warning "vcpkg requested but VCPKG_ROOT not set"
    } else {
        Write-Host "Using system packages (vcpkg not used)" -ForegroundColor Cyan
    }
    
    # Configure
    Write-Host "Configuring project..." -ForegroundColor Yellow
    & cmake @cmakeArgs
    
    if ($LASTEXITCODE -ne 0) {
        throw "CMake configuration failed"
    }
    
    # Build
    Write-Host "Building project..." -ForegroundColor Yellow
    & cmake --build . --config $BuildType
    
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }
    
    Write-Host ""
    Write-Host "Build completed successfully!" -ForegroundColor Green
    Write-Host "Executable location: build\$BuildType\wtop.exe" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage examples:" -ForegroundColor Yellow
    Write-Host "  .\$BuildType\wtop.exe --help          Show help"
    Write-Host "  .\$BuildType\wtop.exe                 Start with default settings"
    Write-Host "  .\$BuildType\wtop.exe --output json   JSON output mode"
    Write-Host "  .\$BuildType\wtop.exe --refresh 500   Custom refresh rate"
    
} catch {
    Write-Error "Build failed: $_"
    exit 1
} finally {
    Set-Location ".."
}
