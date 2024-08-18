# tinyenv

A tiny replacement of `*env` (rbenv, plenv, goenv, ....)

A major difference from `*env` is that tinyenv does NOT provide installers of languages.
You should install langauages manually. See [LANGUAGE-INSTALL.md](LANGUAGE-INSTALL.md).

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

# install java from https://adoptium.net/temurin/releases/
# linux
❯ curl -fsSL https://api.adoptium.net/v3/binary/latest/17/ga/linux/x64/jdk/hotspot/normal/eclipse | tar xzf - -C ~/.tinyenv/java/versions

❯ tinyenv java versions
jdk-17.0.12+7

❯ tinyenv java global jdk-17.0.12+7

❯ java --version
openjdk 17.0.12 2024-07-16
OpenJDK Runtime Environment Temurin-17.0.12+7 (build 17.0.12+7)
OpenJDK 64-Bit Server VM Temurin-17.0.12+7 (build 17.0.12+7, mixed mode, sharing)
```

# Author

Shoichi Kaji

# License

MIT
