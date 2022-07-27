# 🐬🧋 FlipperZero-Tea
[![lint](https://github.com/jon4hz/flipperzero-tea/actions/workflows/lint.yml/badge.svg)](https://github.com/jon4hz/flipperzero-tea/actions/workflows/lint.yml)
[![goreleaser](https://github.com/jon4hz/flipperzero-tea/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/jon4hz/flipperzero-tea/actions/workflows/goreleaser.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jon4hz/flipperzero-tea)](https://goreportcard.com/report/github.com/jon4hz/flipperzero-tea)

A [bubbletea](https://github.com/charmbracelet/bubbletea)-bubble and TUI to interact with your flipper zero.  
The flipper will be automatically detected, if multiple flippers are connected, the first one will be used.

## 🚀 Installation
```
go install github.com/jon4hz/flipperzero-tea@latest
```

## ✨ Usage
```bash
# trying to autodetect that dolphin
$ flipperzero-tea

# no flipper found automatically :(
$ flipperzero-tea -p /dev/ttyACM0
```

## ⚡️ SSH
Flipperzero-tea also allows you to start an ssh server, serving the flipper zero ui over a remote connection.  
Why? - Why not!
```bash
# start the ssh server listening on localhost:2222 (default)
$ flipperzero-tea server -l 127.0.0.1:2222

# connect to the server (from the same machine)
$ ssh localhost -p 2222
```