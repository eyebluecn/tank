#!/bin/bash

###########################################################################
#
#  Tank build script for Linux or MacOS
#
###########################################################################

#prepare the variables.

# version name
VERSION_NAME=tank-3.0.4
echo "VERSION_NAME: ${VERSION_NAME}"
#  golang proxy
GOPROXY=https://athens.azurefd.net
echo "GOPROXY: ${GOPROXY}"
# eg. amd64
GOARCH=$(go env GOARCH)
echo "GOARCH: ${GOARCH}"
# eg. /data/golang
GOPATH=$(go env GOPATH)
echo "GOPATH: ${GOPATH}"
# eg. darwin
GOOS=$(go env GOOS)
echo "GOOS: ${GOOS}"
# service dir eg. /data/tank/build/pack
PACK_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo "PACK_DIR: ${PACK_DIR}"
# build dir eg. /data/tank/build
BUILD_DIR=$(dirname ${PACK_DIR})
echo "BUILD_DIR: ${BUILD_DIR}"
# project dir eg. /data/tank
PROJECT_DIR=$(dirname ${BUILD_DIR})
echo "PROJECT_DIR: ${PROJECT_DIR}"
# final zip file name.
FILE_NAME=${VERSION_NAME}.${GOOS}-${GOARCH}.tar.gz
echo "FILE_NAME: ${FILE_NAME}"
# zip dist dir eg. /data/tank/tmp/dist
DIST_DIR=${PROJECT_DIR}/tmp/dist
echo "DIST_DIR: ${DIST_DIR}"
# component dir eg. /data/tank/tmp/dist/tank-x.x.x
COMPONENT_DIR=${DIST_DIR}/${VERSION_NAME}
echo "COMPONENT_DIR: ${COMPONENT_DIR}"
# final dist path eg. /data/tank/tmp/dist/tank-x.x.x.darwin-amd64.tar.gz
DIST_PATH=${DIST_DIR}/${FILE_NAME}
echo "DIST_PATH: ${DIST_PATH}"

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