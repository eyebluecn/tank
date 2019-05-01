#!/bin/bash

#prepare the variables.

# version name
VERSION_NAME=3.0.0-beta.1
# eg. amd64
GOARCH=$(go env GOARCH)
# eg. darwin
GOOS=$(go env GOOS)
# service dir eg. /data/tank/build/service
SERVICE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# project dir eg. /data/tank
PROJECT_DIR=$(dirname $(dirname ${SERVICE_DIR}))
# final zip file name.
FILE_NAME=${VERSION_NAME}.${GOOS}-${GOARCH}.tar.gz
# zip dist dir
DIST_DIR=${PROJECT_DIR}/tmp/dist
# final dist path
DIST_PATH=${DIST_DIR}/${FILE_NAME}

cd ${PROJECT_DIR}

echo "go build -mod=readonly"
go build -mod=readonly

# if a directory
if [[ ! -d DIST_DIR ]] ; then
    mkdir -p ${DIST_DIR}
fi

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
echo "tar -zcvf $distFolder/$FILE_NAME ./$VERSION_NAME"
cd $distPath
cd ..
tar -zcvf $distFolder/$FILE_NAME ./$VERSION_NAME

echo "check the dist file in $distPath"
echo "finish!"