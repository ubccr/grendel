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
#
# $ yum install lzma-sdk-devel xz-devel
#
#
# Credits:
#
# Adopted from: https://github.com/danderson/netboot/blob/master/third_party/Makefile
#

all:
	$(error Please specifcy a target: build, build-with-ca, bindata, clean)

.PHONY: build
build:
	(cd ipxe/src &&\
	sed -i '/^\/\/#define VLAN_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define DIGEST_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define NSLOOKUP_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define PING_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define NTP_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define\tCONSOLE_SYSLOG/s/^\/\///' config/console.h &&\
	sed -i '/^\/\/#define\tCONSOLE_SYSLOGS/s/^\/\///' config/console.h &&\
	make bin/ipxe.pxe \
		bin/undionly.kpxe \
		bin-x86_64-efi/ipxe.efi \
		bin-x86_64-efi/snponly.efi \
		bin-i386-efi/ipxe.efi \
		EMBED=../../boot.ipxe)

.PHONY: build-with-ca
build-with-ca:
	(cd ipxe/src &&\
	sed -i 's/^#undef\tDOWNLOAD_PROTO_HTTPS/#define\tDOWNLOAD_PROTO_HTTPS/' config/general.h &&\
	sed -i '/^\/\/#define IMAGE_TRUST_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define CERT_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define VLAN_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define DIGEST_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define NSLOOKUP_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define PING_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define NTP_CMD/s/^\/\///' config/general.h &&\
	sed -i '/^\/\/#define\tCONSOLE_SYSLOG/s/^\/\///' config/console.h &&\
	sed -i '/^\/\/#define\tCONSOLE_SYSLOGS/s/^\/\///' config/console.h &&\
	make bin/ipxe.pxe \
		bin/undionly.kpxe \
		bin-x86_64-efi/ipxe.efi \
		bin-x86_64-efi/snponly.efi \
		bin-i386-efi/ipxe.efi \
		CERT=../../../out/GrendelCA.crt TRUST=../../../out/GrendelCA.crt \
		EMBED=../../boot.ipxe)

.PHONY: bindata
bindata:
	(rm -rf bin && mkdir bin)
	mv -f ipxe/src/bin/ipxe.pxe bin/ipxe.pxe
	mv -f ipxe/src/bin/undionly.kpxe bin/undionly.kpxe
	mv -f ipxe/src/bin-x86_64-efi/ipxe.efi bin/ipxe-x86_64.efi
	mv -f ipxe/src/bin-x86_64-efi/snponly.efi bin/snponly-x86_64.efi
	mv -f ipxe/src/bin-i386-efi/ipxe.efi bin/ipxe-i386.efi

.PHONY: clean
clean:
	rm -rf bin/
	(cd ipxe/src && git checkout -- config/general.h && git checkout -- config/console.h && make veryclean)
