package main

import (
	"bytes"
	"log"
	"os"

	"github.com/makoto126/term-atari/gui"
	"github.com/makoto126/term-atari/vm"
)

func main() {

	term := new(gui.Term)
	if err := term.Init(); err != nil {
		log.Fatalln(err)
	}

	for {
		game := term.GameSelect(AssetNames())
		if game == "" {
			log.Println("no game selected")
			os.Exit(0)
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
			term.Quit(),
		)

		if err := chip8.Load(bytes.NewBuffer(data)); err != nil {
			log.Fatalln(err)
		}

		if err := chip8.Loop(); err != nil {
			log.Fatalln(err)
		}
	}
}
