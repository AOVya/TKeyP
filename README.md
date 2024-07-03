# TkeyP

TkeyP is a simple tool written in go to keylog a keyboard on one machine and simulate the keystrokes on another machine. This is very much still in development.

## Installation

```bash
git clone github.com/aovya/tkeyp
cd tkeyp
go build
```

## Usage

Because this program interacts with the keyboard, it must be run as root. There need to be a sender and a reciever. The sender logs the keystrokes and the reciever simulates the keystrokes.

```bash
sudo ./tkeyp [sender [ip address] [port]| reciever] 
```

Since the reciever just attaces to localhost:6094 there is no need to specify an ip address or port.

## TODOs:

* Support events other than key presses
* Add tests
