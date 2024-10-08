# 🐬🧋 Fztea
[![lint](https://github.com/jon4hz/fztea/actions/workflows/lint.yml/badge.svg)](https://github.com/jon4hz/fztea/actions/workflows/lint.yml)
[![goreleaser](https://github.com/jon4hz/fztea/actions/workflows/release.yml/badge.svg)](https://github.com/jon4hz/fztea/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jon4hz/fztea)](https://goreportcard.com/report/github.com/jon4hz/fztea)
[![Powered by Dolphines](https://img.shields.io/badge/Powered%20by-Dolphins-blue)](https://img.shields.io/badge/Powered%20by-Dolphins-blue)

A [bubbletea](https://github.com/charmbracelet/bubbletea)-bubble and TUI to interact with your [flipper zero](https://flipperzero.one/).  
The flipper will be automatically detected, if multiple flippers are connected, the first one will be used.

## 🚀 Installation
```bash
# using go directly
$ go install github.com/jon4hz/fztea@latest

# from aur (btw)
$ yay -S fztea-bin

# local pkg manager
## debian / ubuntu
$ dpkg -i fztea-v0.6.2-linux-amd64.deb

## rhel / fedora / suse
$ rpm -i fztea-v0.6.2-linux-amd64.rpm

## alpine
$ apk add --allow-untrusted fztea-v0.6.2-linux-amd64.apk

# homebrew (macOS & linux)
$ brew install jon4hz/homebrew-tap/fztea

# windows
# -> I'm sure you'll figure something out :)
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

By default, `fztea` doesn't require any authentication but you can specify an `authorized_keys` file if you want to.

```bash
# use authorized_keys for authentication
$ fztea server -l 127.0.0.1:2222 -k ~/.ssh/authorized_keys
```

## 📸 Screenshots
You can take a screenshot of the flipper using `ctrl+s` at any time. `Fztea` will store the screenshot in the working directoy, by default in a 1024x512px resolution.  
The size of the screenshot can be customized using the `--screenshot-resolution` flag. 
```
$ fztea --screenshot-resolution=1920x1080
```

## ⌨️ Button Mapping
| Key             | Flipper Event | Keypress Type
|-----------------|---------------|--------------|
| w, ↑            | up            | short        |
| d, →            | right         | short        |
| s, ↓            | down          | short        |
| a, ←            | left          | short        |
| o, enter, space | ok            | short        |
| b, back, esc    | back          | short        |
| W, shift + ↑    | up            | long         |
| D, shift + →    | right         | long         |
| S, shift + ↓    | down          | long         |
| A, shift + ←    | left          | long         |
| O               | ok            | long         |
| B               | back          | long         |


## 🌈 Custom colors 
You can set custom fore- and background colors using the `--bg-color` and `--fg-color` flags.
```
$ fztea --bg-color="#8A0000" --fg-color="#000000"
```
Results in:

![ColorScreenshot](/.github/assets/custom_colors.png)



## 🎬 Demo

### Local
![LocalDemo](/.github/assets/demo.gif)
### SSH
https://user-images.githubusercontent.com/26183582/181772189-13d7aeaa-ac26-4701-8104-a71ed218539c.mp4

