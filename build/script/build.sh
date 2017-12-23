#!/bin/bash

# if GOPATH not set
if [ -z "$GOPATH" ] ; then
  echo "GOPATH not defined"
  exit 1
fi


PRE_DIR=$(pwd)

cd $GOPATH

echo "golang.org . Please download from: https://github.com/MXi4oyu/golang.org and put in the directory with same level of github.com"
# echo "go get golang.org/x"
# go get golang.org/x

# resize image
echo "go get github.com/disintegration/imaging"
go get github.com/disintegration/imaging

# json parser
echo "go get github.com/json-iterator/go"
go get github.com/json-iterator/go

# mysql
echo "go get github.com/go-sql-driver/mysql"
go get github.com/go-sql-driver/mysql

# dao database
echo "go get github.com/jinzhu/gorm"
go get github.com/jinzhu/gorm

# uuid
echo "go get github.com/nu7hatch/gouuid"
go get github.com/nu7hatch/gouuid

echo "build tank ..."
go install tank

echo "packaging..."
distPath="$GOPATH/src/tank/dist"

# if a directory
if [ ! -d $distPath ] ; then
    echo "clear $distPath"
    rm -rf $distPath
fi

echo "create directory $distPath"
mkdir $distPath

echo "copying cmd tank"
cp "$GOPATH/bin/tank" $distPath

echo "copying build"
cp -r "$GOPATH/src/tank/build" $distPath

cd $PRE_DIR

echo "check the dist file in $distPath"
echo "finish!"