@echo off
SETLOCAL ENABLEEXTENSIONS

:: Install Scoop if not present
where scoop >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Scoop not found. Installing Scoop...
    powershell -Command "iwr -useb get.scoop.sh | iex"
)

:: Install Go if not present
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Go not found. Installing Go...
    scoop install go
)

:: Install TinyGo if not present
where tinygo >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo TinyGo not found. Installing TinyGo...
    scoop install tinygo
)

:: Install Binaryen if not present
where binaryen-version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Binaryen not found. Installing Binaryen...
    scoop install binaryen
)

:: Install Avrdude if not present
where avrdude >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Avrdude not found. Installing Avrdude...
    scoop install avrdude
)

:: Install BOSSA for Arduino Nano33 IoT if not present
if not exist "c:\Program Files\BOSSA" (
    echo BOSSA not found. Installing BOSSA...
    powershell -Command "Invoke-WebRequest -Uri 'https://github.com/shumatech/BOSSA/releases/download/1.9.1/bossa-x64-1.9.1.msi' -OutFile 'bossa-x64-1.9.1.msi'"
    msiexec /i bossa-x64-1.9.1.msi /quiet
)

echo Installation complete.
