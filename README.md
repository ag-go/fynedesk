<p align="center">
  <a href="https://godoc.org/fyshos.com/fynedesk" title="GoDoc Reference" rel="nofollow"><img src="https://img.shields.io/badge/go-documentation-blue.svg?style=flat" alt="GoDoc Reference"></a>
  <a href="https://github.com/fyshos/fynedesk/releases/tag/v0.4.0" title="0.4.0 Release" rel="nofollow"><img src="https://img.shields.io/badge/version-0.4.0-blue.svg?style=flat" alt="0.4.0 release"></a>
  <a href='http://gophers.slack.com/messages/fynedesk'><img src='https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=blue' alt='Join us on Slack' /></a>

  <br />
  <a href="https://goreportcard.com/report/fyshos.com/fynedesk"><img src="https://goreportcard.com/badge/fyshos.com/fynedesk" alt="Code Status" /></a>
  <a href="https://github.com/fyshos/fynedesk/actions"><img src="https://github.com/fyshos/fynedesk/workflows/Platform%20Tests/badge.svg" alt="Build Status" /></a>
  <a href='https://coveralls.io/github/fyshos/fynedesk?branch=develop'><img src='https://coveralls.io/repos/github/fyshos/fynedesk/badge.svg?branch=develop' alt='Coverage Status' /></a>
</p>

# About FyneDesk

FyneDesk is an easy to use Linux/Unix desktop environment following material design.
It is built using the [Fyne](https://fyne.io) toolkit and is designed to be
easy to use as well as easy to develop. We use the Go language and welcome
any contributions or feedback for the project.

[![FyneDesk v0.4](https://img.youtube.com/vi/82Wu5k0xZOI/0.jpg)](https://www.youtube.com/watch?v=82Wu5k0xZOI)

## Dependencies

### Compiling

Compiling requires the same dependencies as Fyne. See the [Getting Started](https://developer.fyne.io/started/) documentation for installation steps.

### Running

For a full desktop experience you will also need the following external tools installed:

- `arandr` for modifying display settings
- `xbacklight` or `brightnessctl` for laptop brightness
- `connman-gtk` is currently used for configuring Wi-Fi network settings
- `compton` for compositor support

The desktop does work without the runtime dependencies but the experience will be degraded.

## Getting Started

Using standard Go tools you can install FyneDesk using:
```
go get fyshos.com/fynedesk/cmd/fynedesk
```

This will add `fynedesk` to your $GOPATH (usually ~/go/bin).
You can now run the app in "preview" mode like any other Fyne app.
Doing so is not running a window manager, to do so requires another few steps:

### Setting up as a desktop environment

To use this as your main desktop you can run the following commands to set up
fynedesk as a selectable desktop option in your login manager (such as LightDM for example):

```
git clone https://github.com/fyshos/fynedesk
cd fynedesk
make
sudo make install
```

You can now log out and see that it is in your desktop selection list at login.

### Debugging a window manager

You can also run the window manager components in an embedded X window for testing.
You will need the `Xephyr` tool installed for your platform (often installed as part of Xorg).
Once it is present you can use the following command from the same directory as above:

    make embed

It should look like this:

<p align="center" markdown="1">
  <img src="desktop-dark-current.png" alt="Fyne Desktop - Dark" />
</p>

If you run the command when there is a window manager running, or on
an operating system that does not support window managers (Windows or
macOS) then the app will start in UI test mode.
When loaded in this way you can run all of the features except the
controlling of windows - they will load on your main desktop.

## Runner

A desktop needs to be rock solid, and whilst we are working hard to get there,
any alpha or beta software can run into unexpected issues. 
For that reason, we have included a `fynedesk_runner` utility that can help
manage unexpected events. If you start the desktop using the runner, then
if a crash occurs, it will normally recover where it left off with no loss
of data in your applications.

Using standard Go tools you can install the runner using:

    go get fyshos.com/fynedesk/cmd/fynedesk_runner

From then on execute that instead of the `fynedesk` command for a more 
resilient desktop when testing out pre-release builds.

## Design

Design concepts, and the abstract wallpapers have been contributed by [Jost Grant](https://github.com/jostgrant).

## Shipping FyneDesk

If you are installing FyneDesk by default on a distribution, or making it available as a standard option, you should consider the following points.
You do not need to ship the library or any dependencies, but it is recommended to add the following apps as well:

| app | go get | description |
| --- | ------ | ----------- |
| fin | `github.com/fyshos/fin` | A display manager app that matches the look and feel of FyneDesk |

Please do let us know if you package FyneDesk for your system, so we can include a link from here :).
