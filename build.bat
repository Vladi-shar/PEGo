# Windows build (assumes you're running this in a compatible shell, like Git Bash)
go build -ldflags="-H=windowsgui -s -w" -o bin\PEGo.exe .
