#
# This Makefile is used to build ipxe firmware binaries. 
#
# Usage:
# 	make build
#
# Required packages:
# 
# - lzma-sdk-devel
# - xz-devel
# - go-bindata
#
# $ yum install lzma-sdk-devel xz-devel
# $ go get -u github.com/go-bindata/go-bindata/...
#
#
# Credits:
#
# Adopted from: https://github.com/danderson/netboot/blob/master/third_party/Makefile
#

all:
	$(error Please specifcy a target: build)

.PHONY: build
build:
	(cd ipxe/src &&\
	make bin/ipxe.pxe \
		bin/undionly.kpxe \
		bin-x86_64-efi/ipxe.efi \
		bin-x86_64-efi/snponly.efi \
		bin-i386-efi/ipxe.efi \
		EMBED=../../boot.ipxe)

.PHONY: bindata
bindata:
	(rm -rf bin && mkdir bin)
	mv -f ipxe/src/bin/ipxe.pxe bin/ipxe.pxe
	mv -f ipxe/src/bin/undionly.kpxe bin/undionly.kpxe
	mv -f ipxe/src/bin-x86_64-efi/ipxe.efi bin/ipxe-x86_64.efi
	mv -f ipxe/src/bin-x86_64-efi/snponly.efi bin/snponly-x86_64.efi
	mv -f ipxe/src/bin-i386-efi/ipxe.efi bin/ipxe-i386.efi
	go-bindata -o ipxe-bin.go -pkg firmware -nometadata -nomemcopy -prefix bin/ bin/
	gofmt -s -w ipxe-bin.go

.PHONY: clean
clean:
	rm -rf bin/
	(cd ipxe/src && make veryclean)
