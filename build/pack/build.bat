@if "%DEBUG%" == "" echo off
@rem ##########################################################################
@rem
@rem  Tank build script for Windows
@rem
@rem ##########################################################################



@REM ==== START VALIDATION ====
if "%GOPATH%"=="" (
    echo The GOPATH environment variable is not defined correctly
    goto end
)

set PRE_DIR=%cd%

@rem version name
set VERSION_NAME=tank-2.0.0

cd %GOPATH%

@rem echo golang.org . Please download from: https://github.com/eyebluecn/golang.org and put in the directory with same level of github.com
@rem echo go get golang.org/x
@rem go get golang.org/x
echo git clone https://github.com/eyebluecn/golang.org.git %golangOrgFolder%
set golangOrgFolder=%GOPATH%\src\golang.org
if not exist %golangOrgFolder% (
    git clone https://github.com/eyebluecn/golang.org.git %golangOrgFolder%
)

@rem resize image
echo go get github.com/disintegration/imaging
go get github.com/disintegration/imaging

@rem json parser
echo go get github.com/json-iterator/go
go get github.com/json-iterator/go


@rem mysql
echo go get github.com/go-sql-driver/mysql
go get github.com/go-sql-driver/mysql

@rem dao database
echo go get github.com/jinzhu/gorm
go get github.com/jinzhu/gorm


@rem uuid
echo go get github.com/nu7hatch/gouuid
go get github.com/nu7hatch/gouuid

echo build tank ...
go install tank

echo packaging

set distFolder=%GOPATH%\src\tank\dist
if not exist %distFolder% (
    md %distFolder%
)

set distPath=%distFolder%\%VERSION_NAME%
if exist %distPath% (
    echo clear %distPath%
    rmdir /s/q %distPath%
)

echo create directory %distPath%
md %distPath%

echo copying tank.exe
copy %GOPATH%\bin\tank.exe %distPath%

echo copying build
xcopy %GOPATH%\src\tank\build %distPath% /e/h

echo "remove pack"
rmdir /s/q %distPath%\pack

echo "remove service"
rmdir /s/q %distPath%\service

echo "remove doc"
rmdir /s/q %distPath%\doc

cd %PRE_DIR%

echo check the dist file in %distPath%
echo finish!