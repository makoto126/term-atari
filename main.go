package main

import (
	"bytes"
	"log"

	"github.com/makoto126/term-atari/gui"
	"github.com/makoto126/term-atari/vm"
)

func main() {

	for {
		rom := gui.SelectRom(AssetNames())
		if rom == "" {
			break
		}

		term := new(gui.Term)

		quit, err := term.Init()
		if err != nil {
			log.Fatalln(err)
		}

		chip8 := new(vm.Chip8)

		chip8.Init(
			term,
			term,
			term,
			quit,
		)

		data, err := Asset(rom)
		if err != nil {
			log.Fatalln(err)
		}

		if err := chip8.Load(bytes.NewBuffer(data)); err != nil {
			log.Fatalln(err)
		}

		if err := chip8.Loop(); err != nil {
			log.Fatalln(err)
		}
	}
}
