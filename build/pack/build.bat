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

cd %GOPATH%

echo golang.org . Please download from: https://github.com/MXi4oyu/golang.org and put in the directory with same level of github.com
@rem echo go get golang.org/x
@rem go get golang.org/x

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
set distPath=%GOPATH%\src\tank\dist
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

cd %PRE_DIR%

echo check the dist file in %distPath%
echo finish!