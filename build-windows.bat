@echo off
echo Building wtop for Windows...
set GOOS=windows
set GOARCH=amd64
go build -o wtop.exe wtop.go
echo Build complete: wtop.exe
pause
