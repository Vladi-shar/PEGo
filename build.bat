del bin\PEGo.exe

go-winres make -in src\winres\winres.json -out src\rsrc
@REM go-winres make -in src\winres\winres.json


@REM this command builds without console window
go build -ldflags="-H=windowsgui -s -w" -o bin\PEGo.exe .\src

@REM this command builds with console window
@REM go build -ldflags="-s -w" -o bin\PEGo.exe .\src

