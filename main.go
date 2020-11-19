package main

import (
	"bytes"
	"log"

	"github.com/makoto126/term-atari/gui"
	"github.com/makoto126/term-atari/vm"
)

func main() {

	for {
		term := new(gui.Term)
		game, quit, err := term.Init(AssetNames())
		if err != nil {
			log.Fatalln(err)
		}

		data, err := Asset(game)
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

		if err := chip8.Load(bytes.NewBuffer(data)); err != nil {
			log.Fatalln(err)
		}

		if err := chip8.Loop(); err != nil {
			log.Fatalln(err)
		}
	}
}
