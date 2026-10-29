@echo off
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0cleanup-run.ps1"
timeout /t 3 >nul
