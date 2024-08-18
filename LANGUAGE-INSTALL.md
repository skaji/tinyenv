# How to install languages

## Go

https://go.dev/dl/

List
```bash
curl -fsSL 'https://go.dev/dl/?mode=json&include=all' | jq -r '.[] | .version' | sed -e 's/^go//' | head
```

Install
```bash
VERSION=1.23.0
OS=darwin
ARCH=arm64
# OS=linux
# ARCH=amd64
URL=https://dl.google.com/go/go$VERSION.$OS-$ARCH.tar.gz

curl -fsSL -o go.tar.gz $URL
```

## Java

https://adoptium.net/temurin/releases/

Install
```bash
VERSION=17
OS=mac
ARCH=aarch64
# OS=linux
# ARCH=x64
URL=https://api.adoptium.net/v3/binary/latest/$VERSION/ga/$OS/$ARCH/jdk/hotspot/normal/eclipse

curl -fsSL -o jdk.tar.gz $URL
```

NOTE: On macos, jdk.tar.gz contains following files. You may want to use `jdk-17.0.12+7/Contents/Home` only.
```
jdk-17.0.12+7/Contents
jdk-17.0.12+7/Contents/_CodeSignature
jdk-17.0.12+7/Contents/Home
jdk-17.0.12+7/Contents/MacOS
jdk-17.0.12+7/Contents/Info.plist
```

## Node

https://nodejs.org/en/download/prebuilt-binaries

Install
```bash
VERSION=v20.16.0
OS=darwin
ARCH=arm64
# OS=linux
# ARCH=x64
URL=https://nodejs.org/dist/$VERSION/node-$VERSION-$OS-$ARCH.tar.gz

curl -fsSL -o node-$VERSION-$OS-$ARCH.tar.gz $URL
```

## Perl

https://github.com/skaji/relocatable-perl/blob/main/releases.csv

List
```bash
curl -s https://raw.githubusercontent.com/skaji/relocatable-perl/main/releases.csv | perl -F, -anle 'print $F[0] if $. > 1' | uniq | head
```

Install
```bash
VERSION=5.40.0.0
OS=darwin
ARCH=arm64
# OS=linux
# ARCH=amd64
URL=https://github.com/skaji/relocatable-perl/releases/download/$VERSION/perl-$OS-$ARCH.tar.xz

curl -fsSL -o perl-$OS-$ARCH.tar.xz $URL
```

## Python

https://www.python.org/downloads/
https://github.com/indygreg/python-build-standalone/releases
https://gregoryszorc.com/docs/python-build-standalone/main/running.html#obtaining-distributions

Install
```bash
VERSION=3.12.5
DATE=20240814
OS=apple-darwin
ARCH=aarch64
# OS=unknown-linux-gnu
# ARCH=x86_64
URL=https://github.com/indygreg/python-build-standalone/releases/download/$DATE/cpython-$VERSION+$DATE-$ARCH-$OS-install_only.tar.gz

curl -fsSL -o cpython-$VERSION+$DATE-$ARCH-$OS-install_only.tar.gz $URL
```

## Ruby

https://github.com/rbenv/ruby-build

List
```bash
./bin/ruby-build -l
```

Install
```bash
TINYENV_ROOT=$(tinyenv root)
./bin/ruby-build 3.1.6 $TINYENV_ROOT/ruby/versions/3.1.6
```
