# mesh2prod
Deliver the Service Mesh to Production, a parody game using [GOSGE](https://github.com/juan-medina/gosge)

[![License: Apache2](https://img.shields.io/badge/license-Apache%202-blue.svg)](/LICENSE)
[![Build Status](https://travis-ci.com/juan-medina/gosge.svg?branch=main)](https://travis-ci.com/juan-medina/mesh2prod)
[![conduct](https://img.shields.io/badge/code%20of%20conduct-contributor%20covenant%202.0-purple.svg?style=flat-square)](https://www.contributor-covenant.org/version/2/0/code_of_conduct/)

## Info

TBC

## Run the game


```bash
$ make run
```

Alternatively you could run with:

```bash
$ go run main.go
```

## Requirements

### Ubuntu

#### X11

    apt-get install libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev

#### Wayland

    apt-get install libgl1-mesa-dev libwayland-dev libxkbcommon-dev

### Fedora

#### X11

    dnf install mesa-libGL-devel libXi-devel libXcursor-devel libXrandr-devel libXinerama-devel

#### Wayland

    dnf install mesa-libGL-devel wayland-devel libxkbcommon-devel

### macOS

On macOS, you need Xcode or Command Line Tools for Xcode.

### Windows

On Windows, you need C compiler, like [Mingw-w64](https://mingw-w64.org) or [TDM-GCC](http://tdm-gcc.tdragon.net/).
You can also build binary in [MSYS2](https://msys2.github.io/) shell.

## Build Tags

- `opengl21` : uses OpenGL 2.1 backend (default is 3.3)
- `wayland` : builds against Wayland libraries

## Examples Resources
- Gopher Graphics
    - https://awesomeopensource.com/project/egonelbre/gophers
- Game art 2D:
    - https://www.gameart2d.com
-  Mobile Game Graphics
    - https://mobilegamegraphics.com/product/free-parallax-backgrounds
- Of Far Different Nature Music Loops
    - https://fardifferent.itch.io/loops
- freesound.org
    - https://freesound.org/people/SKKreativ/sounds/456255/