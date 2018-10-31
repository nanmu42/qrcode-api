# QR Code API

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

# Build

You need have Zbar library installed, whose details can be found at `README.md` in project root.

```bash
./build.sh
```

# Run

```bash
cp config_example.toml config.toml
# after editing config.toml per your need
./run.sh
```