@REM this command builds without console window
@REM go build -ldflags="-H=windowsgui -s -w" -o bin\PEGo.exe .
go-winres make -in src\winres\winres.json
go build -ldflags="-s -w" -o bin\PEGo.exe .\src

