package gui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

//SelectRom show a menu for rom select
func SelectRom(romList []string) string {

	s, err := tcell.NewScreen()
	if err != nil {
		return ""
	}
	defer s.Fini()

	if err := s.Init(); err != nil {
		return ""
	}

	s.SetStyle(
		tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack),
	)

	sort.Strings(romList)

	selected := 0
	move := make(chan struct{})
	done := make(chan struct{})
	go func() {
		move <- struct{}{}

		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape:
					selected = -1
					close(done)
					return
				case tcell.KeyEnter:
					close(done)
					return
				case tcell.KeyUp, tcell.KeyRight:
					if selected > 0 {
						selected--
						move <- struct{}{}
					}
				case tcell.KeyDown, tcell.KeyLeft:
					if selected < len(romList)-1 {
						selected++
						move <- struct{}{}
					}
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

outer:
	for {
		select {
		case <-done:
			close(move)
			break outer
		case <-move:
			for i, name := range romList {
				s.SetContent(0, i, 0,
					[]rune(strings.Split(name, "/")[1]),
					tcell.StyleDefault.Reverse(i == selected),
				)
			}

			s.SetContent(0, len(romList)+1, 0,
				[]rune("Keys for Menu: ↑，↓，ESC, ENTER"),
				tcell.StyleDefault.Foreground(tcell.ColorGreen),
			)
			s.SetContent(0, len(romList)+2, 0,
				[]rune("Keys for Game: 1，2，3, 4, q, w, e, r, a, s, d, f, z, x, c, v"),
				tcell.StyleDefault.Foreground(tcell.ColorGreen),
			)

			s.Show()
		}
	}

	if selected == -1 {
		return ""
	}

	return romList[selected]
}
