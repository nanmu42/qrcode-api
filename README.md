# QR Code API

[![Go Report Card](https://goreportcard.com/badge/github.com/nanmu42/qrcode-api)](https://goreportcard.com/report/github.com/nanmu42/qrcode-api)
[![GoDoc](https://godoc.org/github.com/nanmu42/qrcode-api?status.svg)](https://godoc.org/github.com/nanmu42/qrcode-api)

A simple API service for QR Code generation/recognition.

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

* HTTP status 500

Something unexpected happened.

# Build and Run

Install and compile ZBar:

```bash
wget https://downloads.sourceforge.net/project/zbar/zbar/0.10/zbar-0.10.tar.bz2
tar -xf zbar-0.10.tar.bz2
cd zbar-0.10
export CFLAGS=""
./configure --disable-video --without-imagemagick --without-qt --without-python --without-gtk --without-x --disable-pthread
make install
```

Go to `cmd/api` or `cmd/bearychat`. More details are in README.md there.

# License

Copyright (c) 2018 LI Zhennan

Use of this work is governed by an MIT License.
You may find a license copy in project root.