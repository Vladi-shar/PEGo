@REM this command builds without console window
@REM go build -ldflags="-H=windowsgui -s -w" -o bin\PEGo.exe .
go build -ldflags="-s -w" -o bin\PEGo.exe .\src

