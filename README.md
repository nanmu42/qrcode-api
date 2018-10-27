# qrcode-api
A simple API service for QR Code generation/recognition

# Compile

Install and compile ZBar:

```bash
wget https://downloads.sourceforge.net/project/zbar/zbar/0.10/zbar-0.10.tar.bz2
tar -xf zbar-0.10.tar.bz2
cd zbar-0.10
export CFLAGS=""
./configure --disable-video --without-imagemagick --without-qt --without-python --without-gtk --without-x --disable-pthread
make install
```