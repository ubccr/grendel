FROM alpine:3.22 AS build

# the ca bundle is for pulling boot images and talking to BMCs over https, and comes with the alpine base image
RUN mkdir -p /out/etc/ssl/certs /out/etc/grendel /out/tmp /out/var/lib/grendel/images /out/var/lib/grendel/repo /out/var/lib/grendel/templates /out/var/lib/grendel/db \
    && chmod 1777 /out/tmp \
    && cp /etc/ssl/certs/ca-certificates.crt /out/etc/ssl/certs/

# the binaries are built with osusergo, so a uid is resolved by reading this rather than through libc
RUN echo 'root:x:0:0:root:/root:/sbin/nologin' > /out/etc/passwd && echo 'root:x:0:' > /out/etc/group

COPY configs/grendel.toml.sample /out/etc/grendel/grendel.toml


FROM scratch

ARG TARGETOS
ARG TARGETARCH

COPY --from=build /out/ /
COPY ${TARGETOS}/${TARGETARCH}/grendeld /usr/bin/grendeld
COPY ${TARGETOS}/${TARGETARCH}/grendel /usr/bin/grendel

ENV HOME=/root

WORKDIR /var/lib/grendel

EXPOSE 80/tcp 8080/tcp 53/tcp 53/udp 67/udp 69/udp 4011/udp

ENTRYPOINT ["/usr/bin/grendeld"]

CMD ["serve", "--verbose"]
