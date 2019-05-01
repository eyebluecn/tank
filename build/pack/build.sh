#!/bin/bash

#prepare the variables.

# version name
VERSION_NAME=tank-3.0.0.beta1
# eg. amd64
GOARCH=$(go env GOARCH)
# eg. darwin
GOOS=$(go env GOOS)
# service dir eg. /data/tank/build/pack
SERVICE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# project dir eg. /data/tank
PROJECT_DIR=$(dirname $(dirname ${SERVICE_DIR}))
# build dir
BUILD_DIR=${PROJECT_DIR}/build
# final zip file name.
FILE_NAME=${VERSION_NAME}.${GOOS}-${GOARCH}.tar.gz
# zip dist dir eg. /data/tank/tmp/dist
DIST_DIR=${PROJECT_DIR}/tmp/dist
# component dir eg. /data/tank/tmp/dist/tank-3.0.0.beta1
COMPONENT_DIR=${DIST_DIR}/${VERSION_NAME}
# final dist path eg. /data/tank/tmp/dist/tank-3.0.0.beta1.darwin-amd64.tar.gz
DIST_PATH=${DIST_DIR}/${FILE_NAME}

cd ${PROJECT_DIR}

echo "go build -mod=readonly"
go build -mod=readonly

# if a directory
if [[ -d COMPONENT_DIR ]] ; then
    rm -rf ${COMPONENT_DIR}
    mkdir ${COMPONENT_DIR}
else
    mkdir -p ${COMPONENT_DIR}
fi

echo "copying cmd tank"
cp ./tank ${COMPONENT_DIR}

echo "copying build"
cp -r ${BUILD_DIR}/* ${COMPONENT_DIR}

echo "remove pack"
rm -rf ${COMPONENT_DIR}/pack

echo "remove doc"
rm -rf ${COMPONENT_DIR}/doc

echo "compress to tar.gz"
echo "tar -zcvf $DIST_PATH $COMPONENT_DIR"

cd ${DIST_DIR}
tar -zcvf ${DIST_PATH} ./${VERSION_NAME}

echo "finish packaging!"