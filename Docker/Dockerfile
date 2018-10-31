FROM ubuntu:xenial as builder

RUN set -eux; \
        apt update; \
        apt install -y \
            git \
            build-essential \
            wget \
            unzip \
    	; \
    	wget -q -O zbar-src.zip "https://github.com/nanmu42/zbar-utf8/archive/master.zip"; \
        unzip -q zbar-src.zip; \
        cd zbar-utf8-master; \
        export CFLAGS=""; \
        ./configure --disable-video --without-imagemagick --without-qt --without-python --without-gtk --without-x --disable-pthread; \
        make install

RUN set -eux; \
        cd /; \
        wget -q "https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz"; \
        tar -C /usr/local -xzf go1.11.1.linux-amd64.tar.gz; \
        export PATH=$PATH:/usr/local/go/bin; \
        cd /; \
        git clone https://github.com/nanmu42/qrcode-api.git; \
        cd qrcode-api; \
        go get; \
        cd cmd/api; \
        ./build.sh; \
        cd ../bearychat; \
        ./build.sh

FROM ubuntu:xenial

ENV LD_LIBRARY_PATH /usr/local/lib

RUN set -eux; \
    apt update; \
    apt install -y ca-certificates; \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    mkdir -p "/usr/local/include"; \
    mkdir -p "/usr/local/lib/pkgconfig"; \
    mkdir -p "/usr/local/include/zbar"; \
    mkdir -p "/qrcode/logs"; \
    mkdir -p "/qrcode/configs"

COPY --from=builder /usr/local/include/zbar.h /usr/local/include/
COPY --from=builder /usr/local/lib/libzbar* /usr/local/lib/
COPY --from=builder /usr/local/lib/pkgconfig/* /usr/local/lib/pkgconfig/
COPY --from=builder /usr/local/include/zbar/* /usr/local/include/zbar/
COPY --from=builder /qrcode-api/cmd/api/qrcode-api /qrcode/
COPY --from=builder /qrcode-api/cmd/bearychat/qrcode-bot /qrcode/

RUN chmod -R +x /qrcode/

WORKDIR /qrcode

