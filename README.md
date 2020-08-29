[![Build Status](https://travis-ci.org/rholder/i3status-title-on-bar.svg?branch=master)](https://travis-ci.org/rholder/i3status-title-on-bar)
[![Latest Version](https://img.shields.io/github/v/release/rholder/i3status-title-on-bar?color=bright-green&sort=semver)](https://github.com/rholder/i3status-title-on-bar/releases/latest)
[![License](https://img.shields.io/badge/license-apache%202.0-brightgreen.svg)](https://github.com/rholder/i3status-title-on-bar/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/rholder/i3status-title-on-bar)](https://goreportcard.com/report/github.com/rholder/i3status-title-on-bar)
# i3status-title-on-bar

Use `i3status-title-on-bar` to inject the active window title into the output of `i3status` as soon as it is updated. 

## Features
* Supports`i3bar` JSON output format from `i3status`
* Adds active window title information into normal `i3status` output
* Detects when the active window title information changes and signals the `i3status` process to update immediately
* Customize the color, width, and position of the window title information to display

## Installation
Release binaries are available for `linux/amd64`, `linux/arm` (v5), and `linux/arm64`. Open an issue if there is interest in binaries for other platforms.

### Linux
Drop the binary into your path, such as `/usr/local/bin`:
```bash
sudo curl -o /usr/local/bin/i3status-title-on-bar -L "https://github.com/rholder/i3status-title-on-bar/releases/download/v0.6.1/i3status-title-on-bar-linux_amd64" && \
sudo chmod +x /usr/local/bin/i3status-title-on-bar
```

## Usage
By default, the i3 bar configuration usually looks something like this (with `i3status` set as the process that periodically generates i3 bar formatted JSON):
```
bar {
        status_command i3status
}
```

This bit of configuration is where we'll add `i3status-title-on-bar`. Here is how to pipe the i3 bar formatted output from `i3status` into `i3status-title-on-bar` with a custom color:
```
bar {
        status_command i3status | i3status-title-on-bar --color '#02FFBF'
}
```

Here is the full usage of `i3status-title-on-bar` from `--help`:
```
Usage: i3status-title-on-bar [OPTIONS...]

  Use i3status-title-on-bar to prepend the currently active X11 window title
  to the beginning (left) of the i3status output JSON as a new node. From there,
  i3status should be able to pick it up and display it on the bar.

Options:
  --color [i3_color_code]  Set the text color of the JSON node (Defaults to #00FF00)
  --append-end             Append window title JSON node to the end instead of the beginning
  --fixed-width [integer]  Truncate and pad to a fixed width, useful with append-end
  --help                   Print this help text and exit
  --version                Print the version and exit

Examples:
  i3status | i3status-title-on-bar --color '#00EE00'
  i3status | i3status-title-on-bar --append-end --fixed-width 64
  i3status-title-on-bar < i3status-output-example.json

Report bugs and find the latest updates at https://github.com/rholder/i3status-title-on-bar.
```

## Background
Because `i3status` relies on a user configurable polling mechanism (intentionally, to reduce unnecessary system calls) when generating content for the i3 bar, it needs to be notified that an update should occur sooner than the next scheduled wakeup. Without notification, adding the window title to the produced JSON from `i3status` has a variable delay in displaying that depends on the polling interval. This is most noticeable when switching tabs in a browser or text editor where the window title changes based on the active tab but the update to the window title doesn't happen immediately and instead appears to lag behind until `i3status` finally wakes up. Here is a crude diagram of how `i3status-title-on-bar` is affected by `i3status`'s sleep:
```
i3status wakes up, outputs JSON ---> i3status-title-on-bar gets active window title, adds to JSON
       ^                                                           |
       |                                                           v
i3status sleeps <------------------------------------ i3wm displays window title
```

With the addition of [this commit](https://github.com/i3/i3status/commit/0a608d4af67fe59390f2e8931f61b664f48660db), we can force `i3status` to wake up immediately by sending a `USR1` signal to the running `i3status` process id. Thus, we add another subsystem to `i3status-title-on-bar` to listen for window title changes and signal `i3status` when an update occurs. Here is another crude diagram:
```
i3status-title-on-bar detects title change ----> i3status-title-on-bar signals i3status with USR1
                      ^                                                        |
                      |                                                        |
i3status-title-on-bar waits for next change <----------------------------------
```

With these two systems in place, we can reliably update the window title when it changes and display it in the i3 bar.

However, what happens when some process decides it wants to update its own window title constantly all the time triggering constant and very frequent updates to `i3status`? I've attempted to mitigate this behavior by sampling window title changes as they are detected instead of passing them through directly. An update signal to `i3status` is only sent at a max rate of every 100 milliseconds instead of every time a window title property change occurs (that number comes from [here](https://www.nngroup.com/articles/response-times-3-important-limits/)). This minimizes the `USR1` signal sending to `i3status` which forces an update to everything it may be polling.

## Development
Set up a go 1.15 development environment. There are many "valid" or "right" or "idiomatic" ways of doing this. Find the one that works for you that lets you compile and run go code.

Here's what I do to maintain a single isolated project in a single isolated workspace after cloning this git repository:
```
cd i3status-title-on-bar
source .source_me
make
```
An example of the contents of this project's `.source_me` can be found in [.source_me_example](https://github.com/rholder/i3status-title-on-bar/blob/master/.source_me_example). Modify it to suit your needs if you want or use your current go development setup.

The `Makefile` contains a `help` target that displays the following:
```
Usage:

  clean      remove the build directory
  build      assemble the project and place a binary in build/ for this OS
  fmt        run gofmt for the project
  test       run the unit tests
  coverage   run the unit tests with test coverage output to build/coverage.html
  help       prints this help message
```

## License
`i3status-title-on-bar` is released under version 2.0 of the [Apache License](http://www.apache.org/licenses/LICENSE-2.0).
