@echo off
setlocal enabledelayedexpansion

echo Setting up Visual Studio Environment and Building wtop
echo =====================================================

:: Try to find Visual Studio installation
set "VS_PATHS="
set "VS_PATHS=%VS_PATHS% "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat""
set "VS_PATHS=%VS_PATHS% "C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\Build\vcvars64.bat""
set "VS_PATHS=%VS_PATHS% "C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Auxiliary\Build\vcvars64.bat""
set "VS_PATHS=%VS_PATHS% "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvars64.bat""
set "VS_PATHS=%VS_PATHS% "C:\Program Files (x86)\Microsoft Visual Studio\2019\Professional\VC\Auxiliary\Build\vcvars64.bat""
set "VS_PATHS=%VS_PATHS% "C:\Program Files (x86)\Microsoft Visual Studio\2019\Enterprise\VC\Auxiliary\Build\vcvars64.bat""

set "VCVARS_FOUND="
for %%p in (%VS_PATHS%) do (
    if exist %%p (
        set "VCVARS_FOUND=%%p"
        goto :found_vcvars
    )
)

:found_vcvars
if not defined VCVARS_FOUND (
    echo Error: Could not find Visual Studio installation.
    echo Please install Visual Studio 2019 or 2022 with C++ development tools.
    echo.
    echo You can download it from:
    echo https://visualstudio.microsoft.com/downloads/
    echo.
    echo Make sure to install "Desktop development with C++" workload.
    pause
    exit /b 1
)

echo Found Visual Studio at: %VCVARS_FOUND%
echo Setting up environment...

:: Call vcvars to set up the environment
call %VCVARS_FOUND%

if errorlevel 1 (
    echo Error: Failed to set up Visual Studio environment
    exit /b 1
)

echo Environment set up successfully!
echo.

:: Now call the regular build script
call build.bat %*
