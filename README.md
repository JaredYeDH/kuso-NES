# Introduction

The **N**intendo **E**ntertainment **S**ystem (NES) was the world’s most widely used videogames console during the 1980s. From its initial release in 1983 until it was discontinued in 1995, the console brought gaming into more homes than ever before and paved the way for the videogame industry as it stands today. 

Although technology has improved dramatically since the NES, many excellent games were only released on that format and so are unplayable on more modern systems. However these games have been able to survive and continue to be played thanks to **emulation**, which simulates the workings of one system in order to allow software created for it to be used on a modern system. Here is one simple NES emulator written in [Golang](http://golang.org).

Now I've achieved my goal. This project won't get updated anymore.

# Usage

For *nix users, use the following instruction:

```bash
kuso-NES <your .nes/.zip file path>
```

For windows users, use the same instruction is OK. But the easiest way is just drag the rom file and drop to the kuso-NES.exe. Then it will run automaticly.

# Key Map

| Keyboard | NES Controller     |
| -------- | ------------------ |
| W,S,A,D  | Up,Down,Left,Right |
| K        | A                  |
| J        | B                  |
| F        | Select             |
| H        | Start              |

# Installation

Just install the dependencies and run

``` bash
go get github.com/kuso-kodo/kuso-NES
```

and have a coffee. Then every thing is OK.

# Dependencies

[go-gl/gl](https://github.com/go-gl/gl)

[go-gl/glfw](https://github.com/go-gl/glfw)

[go bindings for portaudio](https://github.com/gordonklaus/portaudio)

[portaudio](http://www.portaudio.com/)

# Document

If you want to build a simple NES emulator or just want to find out how NES worked.You can visit [neadev](http://nesdev.com/) for more information.
