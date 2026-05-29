@echo off
rem ===============================================================
rem  NEXTonebotfilter - build script (Windows)
rem  Produces a single self-contained nextonebotfilter.exe with the
rem  Next.js console embedded. Run this once after pulling, or
rem  whenever frontend/backend code changes.
rem ===============================================================
setlocal enabledelayedexpansion
cd /d "%~dp0"

set EMBED_DIR=backend\internal\server\web

echo [1/3] building Next.js console (static export)
if not exist "console\node_modules" (
  pushd console
  call npm install --no-audit --no-fund || goto :fail
  popd
)
pushd console
call npm run build:export || (popd ^& goto :fail)
popd

echo [2/3] copying console/out -^> %EMBED_DIR%
if exist "%EMBED_DIR%" rmdir /s /q "%EMBED_DIR%"
xcopy /E /I /Y /Q console\out "%EMBED_DIR%" >nul || goto :fail

echo [3/3] building Go binary (CGO disabled, pure-Go SQLite)
pushd backend
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w" -o nextonebotfilter.exe ./cmd/nextonebotfilter || (popd ^& goto :fail)
popd

echo.
echo build OK -^> backend\nextonebotfilter.exe
echo run: start.cmd
endlocal & exit /b 0

:fail
echo.
echo [error] build failed. See output above.
echo Press any key to close...
pause >nul
endlocal & exit /b 1
