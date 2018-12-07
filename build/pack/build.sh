#!/bin/bash

# if GOPATH not set
if [ -z "$GOPATH" ] ; then
  echo "GOPATH not defined"
  exit 1
fi

PRE_DIR=$(pwd)

VERSION_NAME=tank-2.0.0
FINAL_NAME=$VERSION_NAME.linux-amd64.tar.gz

cd $GOPATH

# echo "go get golang.org/x"
# go get golang.org/x
if [ ! -d "$GOPATH/src/golang.org" ] ; then
  echo "git clone https://github.com/eyebluecn/golang.org.git"
  git clone https://github.com/eyebluecn/golang.org.git $GOPATH/src/golang.org
fi

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
distFolder="$GOPATH/src/tank/dist"

# if a directory
if [ ! -d distFolder ] ; then
    mkdir $distFolder
fi

distPath=$distFolder/$VERSION_NAME

# if a directory
if [ -d $distPath ] ; then
    echo "clear $distPath"
    rm -rf $distPath
fi

echo "create directory $distPath"
mkdir $distPath

echo "copying cmd tank"
cp "$GOPATH/bin/tank" $distPath

echo "copying build"
cp -r "$GOPATH/src/tank/build/." $distPath

echo "remove pack"
rm -rf $distPath/pack

echo "remove doc"
rm -rf $distPath/doc

echo "compress to tar.gz"
echo "tar -zcvf $distFolder/$FINAL_NAME ./$VERSION_NAME"
cd $distPath
cd ..
tar -zcvf $distFolder/$FINAL_NAME ./$VERSION_NAME

cd $PRE_DIR

echo "check the dist file in $distPath"
echo "finish!"