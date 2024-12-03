#!/bin/bash

GRENDEL_DIR='./.grendel-release'
VERSION=`git describe --long --tags --dirty --always | sed -e 's/^v//'`
NAME=grendel-${VERSION}-linux-amd64
REL_DIR=${GRENDEL_DIR}/${NAME}

rm -Rf ${GRENDEL_DIR}
mkdir -p ${REL_DIR}

go build -ldflags "-X github.com/ubccr/grendel/internal/api.Version=$VERSION" .
cp ./grendel ${REL_DIR}/ 
cp ./configs/grendel.toml.sample ${REL_DIR}/ 
cp ./README.md ${REL_DIR}/ 
cp ./AUTHORS.md ${REL_DIR}/ 
cp ./CHANGELOG.md ${REL_DIR}/ 
cp ./LICENSE ${REL_DIR}/ 
cp ./NOTICE ${REL_DIR}/ 

tar -C ${GRENDEL_DIR} -cvzf ${NAME}.tar.gz ${NAME}
rm -Rf ${GRENDEL_DIR}
