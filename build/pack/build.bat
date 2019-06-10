@if "%DEBUG%" == "" ECHO off
@REM ##########################################################################
@REM
@REM  Tank build script for Windows
@REM  manual https://ss64.com/nt/
@REM
@REM ##########################################################################



@REM prepare the variables.

@REM  version name
SET VERSION_NAME=tank-3.0.4
ECHO VERSION_NAME: %VERSION_NAME%
@REM  golang proxy
SET GOPROXY=https://athens.azurefd.net
ECHO GOPROXY: %GOPROXY%
@REM  assign variable like Linux GOARCH=$(go env GOARCH) eg. amd64
FOR /f %%i IN ('go env GOARCH') DO SET GOARCH=%%i
ECHO GOARCH: %GOARCH%
@REM  eg. D:\Group\Golang
FOR /f %%i IN ('go env GOPATH') DO SET GOPATH=%%i
ECHO GOPATH: %GOPATH%
@REM  eg. windows
FOR /f %%i IN ('go env GOOS') DO SET GOOS=%%i
ECHO GOOS: %GOOS%
@REM  service dir eg. D:\Group\eyeblue\tank\build\pack
SET PACK_DIR=%CD%
ECHO PACK_DIR: %PACK_DIR%
@REM  build dir eg. D:\Group\eyeblue\tank\build
FOR %%F IN (%CD%) DO SET BUILD_DIR_SLASH=%%~dpF
SET BUILD_DIR=%BUILD_DIR_SLASH:~0,-1%
ECHO BUILD_DIR: %BUILD_DIR%
@REM project dir eg. D:\Group\eyeblue\tank
FOR %%F IN (%BUILD_DIR%) DO SET PROJECT_DIR_SLASH=%%~dpF
SET PROJECT_DIR=%PROJECT_DIR_SLASH:~0,-1%
ECHO PROJECT_DIR: %PROJECT_DIR%

@REM  final zip file name. eg. tank-x.x.x.windows-amd64.zip
SET FILE_NAME=%VERSION_NAME%.%GOOS%-%GOARCH%.zip
ECHO FILE_NAME: %FILE_NAME%
@REM  zip dist dir eg. D:\Group\eyeblue\tank\tmp\dist
SET DIST_DIR=%PROJECT_DIR%\tmp\dist
ECHO DIST_DIR: %DIST_DIR%
@REM  component dir eg. D:\Group\eyeblue\tank\tmp\dist\tank-x.x.x
SET COMPONENT_DIR=%DIST_DIR%\%VERSION_NAME%
ECHO COMPONENT_DIR: %COMPONENT_DIR%
@REM  final dist path eg. D:\Group\eyeblue\tank\tmp\dist\tank-x.x.x.windows-amd64.zip
SET DIST_PATH=%DIST_DIR%\%FILE_NAME%
ECHO DIST_PATH: %DIST_PATH%

cd %PROJECT_DIR%

ECHO go build -mod=readonly
go build -mod=readonly


IF EXIST %COMPONENT_DIR% (
    rmdir /s/q %COMPONENT_DIR%
    md %COMPONENT_DIR%
) ELSE (
    md %COMPONENT_DIR%
)


ECHO copy .\tank.exe %COMPONENT_DIR%
copy .\tank.exe %COMPONENT_DIR%

ECHO %BUILD_DIR%\conf %COMPONENT_DIR%\conf /E/H/I
xcopy %BUILD_DIR%\conf %COMPONENT_DIR%\conf /E/H/I

ECHO %BUILD_DIR%\html %COMPONENT_DIR%\html /E/H/I
xcopy %BUILD_DIR%\html %COMPONENT_DIR%\html /E/H/I

ECHO please zip to %DIST_PATH%

ECHO finish packaging!