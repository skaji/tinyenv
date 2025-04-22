# tinyenv

A tiny replacement of `*env` (rbenv, plenv, goenv, ....)

# Install

```bash
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

```
Usage:
  ❯ tinyenv GLOBAL_COMMAND...
  ❯ tinyenv LANGUAGE COMMAND...

Global Commands:
  latest
  rehash
  root
  version
  versions

Languages:
  go
  java
  node
  perl
  python
  raku
  ruby
  solr

Commands:
  global
  install
  latest
  rehash
  reset
  version
  versions

Examples:
  ❯ tinyenv versions
  ❯ tinyenv python install -l
  ❯ tinyenv python install 3.9.19+20240814
  ❯ tinyenv python install latest
  ❯ tinyenv python global 3.12.5+20240814
```

# Example

```console
❯ tinyenv java install -l
temurin-23.0.2+7
temurin-22.0.2+9
temurin-21.0.6+7
temurin-20.0.2+9
temurin-19.0.2+7
temurin-18.0.2.1+1
temurin-17.0.14+7
temurin-11.0.26+4

❯ tinyenv java install -g latest
---> Downloading https://api.adoptium.net/v3/binary/version/jdk-23.0.2+7/mac/aarch64/jdk/hotspot/normal/eclipse
---> Extracting /Users/skaji/.tinyenv/java/cache/temurin-23.0.2+7.tar.gz

❯ java --version
openjdk 23.0.2 2025-01-21
OpenJDK Runtime Environment Temurin-23.0.2+7 (build 23.0.2+7)
OpenJDK 64-Bit Server VM Temurin-23.0.2+7 (build 23.0.2+7, mixed mode, sharing)
```

# Author

Shoichi Kaji

# License

MIT
