#!/bin/bash
GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o MagnetMonitor.exe main.go