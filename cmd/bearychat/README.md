# Bearychat Integration

`qrcode-api` can integrate with [Bearychat](https://bearychat.com/).

This service relies on a running `qrcode-api` instance.

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