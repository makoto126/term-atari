package gui

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	keyPressInterval = 80 * time.Millisecond
	keyWaitInterval  = 10 * time.Millisecond
)

var keymap = map[rune]byte{
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 12,
	'q': 4,
	'w': 5,
	'e': 6,
	'r': 13,
	'a': 7,
	's': 8,
	'd': 9,
	'f': 14,
	'z': 10,
	'x': 0,
	'c': 11,
	'v': 15,
}

//Term is the atari gui
type Term struct {
	s tcell.Screen

	ek  *tcell.EventKey
	gfx [64][32]bool

	quit chan struct{}
}

//Init the Term
func (t *Term) Init() (<-chan struct{}, error) {

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(
		tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack),
	)

	t.s = s
	t.quit = make(chan struct{})

	go func() {
		for {
			ev := t.s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape:
					t.s.Fini()
					close(t.quit)
					return
				case tcell.KeyRune:
					if _, ok := keymap[ev.Rune()]; ok {
						t.ek = ev
					}
				}
			case *tcell.EventResize:
				t.s.Sync()
			}
		}
	}()

	return t.quit, nil
}

//IsPressed Impl
func (t *Term) IsPressed(b byte) bool {

	if t.ek != nil {
		return b == keymap[t.ek.Rune()] && t.ek.When().Add(keyPressInterval).After(time.Now())
	}

	return false
}

//WaitKey Impl
func (t *Term) WaitKey() byte {

	now := time.Now()
	for range time.Tick(keyWaitInterval) {
		if t.ek != nil && t.ek.When().After(now) {
			break
		}
	}
	return keymap[t.ek.Rune()]
}

//Clear Impl
func (t *Term) Clear() {

	for i := 0; i < 64; i++ {
		for j := 0; j < 32; j++ {
			t.gfx[i][j] = false
		}
	}
	t.s.Clear()
}

//Draw Impl
func (t *Term) Draw(x, y int, mem []byte) byte {

	var flag byte
	var xi, yj int
	for j, m := range mem {
		yj = y + j
		if yj >= 32 {
			break
		}
		for i := 0; i < 8; i++ {
			if m&(0x80>>uint(i)) != 0 {
				xi = x + i
				if xi >= 64 {
					break
				}
				if t.gfx[xi][yj] {
					t.gfx[xi][yj] = false
					flag = 1
				} else {
					t.gfx[xi][yj] = true
				}
				t.fill(xi, yj)
			}
		}
	}

	t.s.Show()
	return flag
}

func (t *Term) fill(i, j int) {

	style := tcell.StyleDefault
	if t.gfx[i][j] {
		style = style.Background(tcell.ColorWhite)
	}

	t.s.SetContent(2*i, j, rune('　'), nil, style)
}

//Beep Impl
func (t *Term) Beep() {
	t.s.Beep()
}
