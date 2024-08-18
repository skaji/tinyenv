# tinyenv

A tiny replacement of `*env` (rbenv, plenv, goenv, ....)

# Install

```console
# (1) create a directory
❯ mkdir -p ~/.tinyenv/bin

# (2) download tinyenv, and locate it in ~/.tinyenv/bin
# mac
❯ curl -fsSL https://github.com/skaji/tinyenv/releases/latest/download/tinyenv-darwin-arm64.tar.gz | tar xzf - -C ~/.tinyenv/bin tinyenv
# linux
❯ curl -fsSL https://github.com/skaji/tinyenv/releases/latest/download/tinyenv-linux-amd64.tar.gz | tar xzf - -C ~/.tinyenv/bin tinyenv

# (3) add PATH and setup zsh-completions
❯ echo "export PATH=$HOME/.tinyenv/bin:$PATH" >> ~/.zshrc
❯ echo 'eval "$(tinyenv zsh-completions)"' >> ~/.zshrc
```

# Usage

```console
❯ tinyenv java init

❯ tinyenv java install -l
22
21
20
19
18
17
16
11
8

❯ tinyenv java install 17
---> Downloading https://api.adoptium.net/v3/binary/latest/17/ga/mac/aarch64/jdk/hotspot/normal/eclipse
---> Extracting /Users/skaji/.tinyenv/java/cache/17-mac-aarch64.tar.gz

❯ tinyenv java global 17

❯ java --version
openjdk 17.0.12 2024-07-16
OpenJDK Runtime Environment Temurin-17.0.12+7 (build 17.0.12+7)
OpenJDK 64-Bit Server VM Temurin-17.0.12+7 (build 17.0.12+7, mixed mode, sharing)
```

# Author

Shoichi Kaji

# License

MIT
