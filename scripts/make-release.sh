#!/bin/bash

GRENDEL_DIR='./.grendel-release'
VERSION=`grep Version api/version.go | egrep -o '[0-9]\.[0-9]\.[0-9]'`
NAME=grendel-${VERSION}-linux-amd64
REL_DIR=${GRENDEL_DIR}/${NAME}

rm -Rf ${GRENDEL_DIR}
mkdir -p ${REL_DIR}

cp ./grendel ${REL_DIR}/ 
cp ./grendel.toml.sample ${REL_DIR}/ 
cp ./README.md ${REL_DIR}/ 
cp ./AUTHORS.md ${REL_DIR}/ 
cp ./CHANGELOG.md ${REL_DIR}/ 
cp ./LICENSE ${REL_DIR}/ 
cp ./NOTICE ${REL_DIR}/ 

tar -C ${GRENDEL_DIR} -cvzf ${NAME}.tar.gz ${NAME}
rm -Rf ${GRENDEL_DIR}
