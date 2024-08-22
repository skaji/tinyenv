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
  version, versions

Languages:
  go, java, node, perl, python, raku, ruby

Commands:
  global, install, reahsh, version, versions

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
jdk-22.0.2+9
jdk-21.0.4+7
jdk-20.0.2+9
jdk-19.0.2+7
jdk-18.0.2.1+1
jdk-17.0.12+7
jdk-11.0.22+7.1

❯ tinyenv java install jdk-21.0.4+7
---> Downloading https://api.adoptium.net/v3/binary/version/jdk-21.0.4+7/mac/aarch64/jdk/hotspot/normal/eclipse
---> Extracting /Users/skaji/src/github.com/skaji/tinyenv/_root/java/cache/jdk-21.0.4+7.tar.gz

❯ tinyenv java global jdk-21.0.4+7

❯ java --version
openjdk 21.0.4 2024-07-16 LTS
OpenJDK Runtime Environment Temurin-21.0.4+7 (build 21.0.4+7-LTS)
OpenJDK 64-Bit Server VM Temurin-21.0.4+7 (build 21.0.4+7-LTS, mixed mode)
```

# Author

Shoichi Kaji

# License

MIT
