# QR Code API

[![Go Report Card](https://goreportcard.com/badge/github.com/nanmu42/qrcode-api)](https://goreportcard.com/report/github.com/nanmu42/qrcode-api)
[![GoDoc](https://godoc.org/github.com/nanmu42/qrcode-api?status.svg)](https://godoc.org/github.com/nanmu42/qrcode-api)
[![Docker Image](https://img.shields.io/badge/Docker-image-blue.svg)](https://hub.docker.com/r/nanmu42/qrcode-api/)

A simple API service for QR Code generation/recognition.

This project provide Bearychat integration.

# API Doc

## Encoding

Request:

```
GET /encode?content=helloWorld&size=400&type=png
```

Params:

* `content` required
* `size` QR Code size in pixel, may not be honored
* `type` `png`(default) or `string`

Response:

* HTTP status 200 OK

A `image/png` or plain text(`type=string`).

* HTTP status 400 Bad Request

Check your params.

* HTTP status 500

Something unexpected happened.

## Decoding

Request:

```
POST /decode
```

Params: image as binary body

Response:

* HTTP status 200 OK

Good decoding:

```json
{
    "ok": true,
    "desc": "",
    "content": [
        "你好"
    ]
}
```

Everything is ok, but nothing recognized:

```json
{
    "ok": true,
    "desc": "",
    "content": null
}
```

Something is wrong:

```json
{
    "ok": false,
    "desc": "file decoding error: image: unknown format",
    "content": null
}
```

* HTTP status 413 Request Entity Too Large

Request Body is too large.

* HTTP status 500

Something unexpected happened.

# Docker Image

There is a [pre-compiled Docker image](https://hub.docker.com/r/nanmu42/qrcode-api/)
alone with C++ dependencies(ZBar), you may pull the image like following:

```bash
docker pull nanmu42/qrcode-api
```

See `Docker` directory for `docker-compose.yaml` and more detail.

# Build and Run

If you'd like to get you hands dirty, you can build this project as following:

Download and compile ZBar for shared dependencies:

```bash
wget https://downloads.sourceforge.net/project/zbar/zbar/0.10/zbar-0.10.tar.bz2
# or, if you are suffering decoding troubles on UTF-8, try this modified version:
# wget https://github.com/nanmu42/zbar-utf8/archive/master.zip
tar -xf zbar-0.10.tar.bz2
cd zbar-0.10
export CFLAGS=""
./configure --disable-video --without-imagemagick --without-qt --without-python --without-gtk --without-x --disable-pthread
make install
```

Go to `cmd/api` or `cmd/bearychat` for further instruction, more details are in README.md there.

# License

Copyright (c) 2018 LI Zhennan

Use of this work is governed by an MIT License.
You may find a license copy in project root.