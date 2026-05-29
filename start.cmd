@echo off
rem ===============================================================
rem  NEXTonebotfilter - one-click launcher (Windows)
rem  Single binary, single window, single log stream.
rem  Run build.cmd first to produce backend\nextonebotfilter.exe.
rem ===============================================================
setlocal enabledelayedexpansion
cd /d "%~dp0"

if "%PORT%"=="" set PORT=8787
set BIN=backend\nextonebotfilter.exe
set LOG=data\nextonebotfilter.log

if not exist "%BIN%" (
  echo [setup] %BIN% not found - running build.cmd first...
  call build.cmd || exit /b 1
)

if not exist data mkdir data

title NEXTonebotfilter :%PORT%
echo ----------------------------------------------------------------
echo  NEXTonebotfilter
echo  console + API : http://localhost:%PORT%
echo  log file      : %LOG%
echo ----------------------------------------------------------------
echo  press Ctrl+C to stop
echo ----------------------------------------------------------------

start "" cmd /c "timeout /t 2 /nobreak >nul & start http://localhost:%PORT%"

"%BIN%" -db data\nextonebotfilter.db -console :%PORT% -log "%LOG%" %*
endlocal & exit /b %errorlevel%
