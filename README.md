# 🐬🧋 Fztea (flipperzero-tea)
[![lint](https://github.com/jon4hz/fztea/actions/workflows/lint.yml/badge.svg)](https://github.com/jon4hz/fztea/actions/workflows/lint.yml)
[![goreleaser](https://github.com/jon4hz/fztea/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/jon4hz/fztea/actions/workflows/goreleaser.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jon4hz/fztea)](https://goreportcard.com/report/github.com/jon4hz/fztea)

A [bubbletea](https://github.com/charmbracelet/bubbletea)-bubble and TUI to interact with your flipper zero.  
The flipper will be automatically detected, if multiple flippers are connected, the first one will be used.

## 🚀 Installation
```bash
# using go directly
$ go install github.com/jon4hz/fztea@latest

# from aur (btw)
$ yay -S fztea-bin

# local pkg manager
## debian / ubuntu
$ dpkg -i fztea-v0.2.0-linux-amd64.deb

## rhel / fedora / suse
$ rpm -i fztea-v0.2.0-linux-amd64.rpm

## alpine
$ apk add --allow-untrusted fztea-v0.2.0-linux-amd64.apk

# windows & macOS
# -> I'm sure you'll figure something out. (No binaries for macOS due to crosscompilation errors)
```

## ✨ Usage
```bash
# trying to autodetect that dolphin
$ fztea

# no flipper found automatically :(
$ fztea -p /dev/ttyACM0
```

## ⚡️ SSH
fztea also allows you to start an ssh server, serving the flipper zero ui over a remote connection.  
Why? - Why not!
```bash
# start the ssh server listening on localhost:2222 (default)
$ fztea server -l 127.0.0.1:2222

# connect to the server (from the same machine)
$ ssh localhost -p 2222
```
Currently the ssh server supports no authentiaction methods so use it with caution. 