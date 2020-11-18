package vm

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

var (
	cpuFreq     = 500
	cpuDuration = time.Duration(1000/cpuFreq) * time.Millisecond

	timerFreq     = 60
	timerDuration = time.Duration(1000/timerFreq) * time.Millisecond

	fontset = []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	funcmap = map[uint16]func(*Chip8){
		//00E0: Clears the screen.
		0x00E0: func(c *Chip8) {
			c.Clear()
			c.pc += 2
		},
		//00EE: Returns from a subroutine.
		0x00EE: func(c *Chip8) {
			c.sp--
			c.pc = c.stack[c.sp] + 2
		},
		//1NNN: Jumps to address NNN.
		0x1000: func(c *Chip8) {
			c.pc = c.getNNN()
		},
		//2NNN: Calls subroutine at NNN.
		0x2000: func(c *Chip8) {
			c.stack[c.sp] = c.pc
			c.sp++
			c.pc = c.getNNN()
		},
		//3XNN: Skips the next instruction if VX equals NN.
		//(Usually the next instruction is a jump to skip a code block)
		0x3000: func(c *Chip8) {
			if c.getVX() == c.getNN() {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//4XNN: Skips the next instruction if VX doesn't equal NN.
		//(Usually the next instruction is a jump to skip a code block)
		0x4000: func(c *Chip8) {
			if c.getVX() != c.getNN() {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//5XY0: Skips the next instruction if VX equals VY.
		//(Usually the next instruction is a jump to skip a code block)
		0x5000: func(c *Chip8) {
			if c.getVX() == c.getVY() {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//6XNN: Sets VX to NN.
		0x6000: func(c *Chip8) {
			c.setVX(c.getNN())
			c.pc += 2
		},
		//7XNN: Adds NN to VX. (Carry flag is not changed)
		0x7000: func(c *Chip8) {
			c.setVX(c.getVX() + c.getNN())
			c.pc += 2
		},
		//8XY0: Sets VX to the value of VY.
		0x8000: func(c *Chip8) {
			c.setVX(c.getVY())
			c.pc += 2
		},
		//8XY1: Sets VX to VX or VY. (Bitwise OR operation)
		0x8001: func(c *Chip8) {
			c.setVX(c.getVX() | c.getVY())
			c.pc += 2
		},
		//8XY2: Sets VX to VX and VY. (Bitwise AND operation)
		0x8002: func(c *Chip8) {
			c.setVX(c.getVX() & c.getVY())
			c.pc += 2
		},
		//8XY3: Sets VX to VX xor VY.
		0x8003: func(c *Chip8) {
			c.setVX(c.getVX() ^ c.getVY())
			c.pc += 2
		},
		//8XY4: Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there isn't.
		0x8004: func(c *Chip8) {
			if 0xFF-c.getVX() < c.getVY() {
				c.setVF(1)
			} else {
				c.setVF(0)
			}
			c.setVX(c.getVX() + c.getVY())
			c.pc += 2
		},
		//8XY5: VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
		0x8005: func(c *Chip8) {
			if c.getVX() < c.getVY() {
				c.setVF(0)
			} else {
				c.setVF(1)
			}
			c.setVX(c.getVX() - c.getVY())
			c.pc += 2
		},
		//8XY6: Stores the least significant bit of VX in VF and then shifts VX to the right by 1.
		0x8006: func(c *Chip8) {
			c.setVF(c.getVX() & 0x01)
			c.setVX(c.getVX() >> 1)
			c.pc += 2
		},
		//8XY7: Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
		0x8007: func(c *Chip8) {
			if c.getVX() > c.getVY() {
				c.setVF(0)
			} else {
				c.setVF(1)
			}
			c.setVX(c.getVY() - c.getVX())
			c.pc += 2
		},
		//8XYE: Stores the most significant bit of VX in VF and then shifts VX to the left by 1.
		0x800E: func(c *Chip8) {
			c.setVF(c.getVX() & 0x80)
			c.setVX(c.getVX() << 1)
			c.pc += 2
		},
		//9XY0: Skips the next instruction if VX doesn't equal VY. (Usually the next instruction is a jump to skip a code block)
		0x9000: func(c *Chip8) {
			if c.getVX() != c.getVY() {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//ANNN: Sets I to the address NNN.
		0xA000: func(c *Chip8) {
			c.index = c.getNNN()
			c.pc += 2
		},
		//BNNN: Jumps to the address NNN plus V0.
		0xB000: func(c *Chip8) {
			c.pc = c.getNNN() + uint16(c.register[0])
		},
		//CXNN: Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN.
		0xC000: func(c *Chip8) {
			c.setVX(byte(rand.Intn(256)) & c.getNN())
			c.pc += 2
		},
		//DXYN: Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
		//Each row of 8 pixels is read as bit-coded starting from memory location I;
		//I value doesn’t change after the execution of this instruction.
		//As described above, VF is set to 1 if any screen pixels are flipped from set to unset
		//when the sprite is drawn, and to 0 if that doesn’t happen
		0xD000: func(c *Chip8) {
			x, y := int(c.getVX()), int(c.getVY())
			h := c.opcode & 0x000F
			c.setVF(c.Draw(x, y, c.mem[c.index:c.index+h]))
			c.pc += 2
		},
		//EX9E: Skips the next instruction if the key stored in VX is pressed. (Usually the next instruction is a jump to skip a code block)
		0xE09E: func(c *Chip8) {
			if c.IsPressed(c.getVX()) {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//EXA1: Skips the next instruction if the key stored in VX isn't pressed. (Usually the next instruction is a jump to skip a code block)
		0xE0A1: func(c *Chip8) {
			if !c.IsPressed(c.getVX()) {
				c.pc += 4
			} else {
				c.pc += 2
			}
		},
		//FX07: Sets VX to the value of the delay timer.
		0xF007: func(c *Chip8) {
			c.setVX(c.delayTimer)
			c.pc += 2
		},
		//FX0A: A key press is awaited, and then stored in VX. (Blocking Operation. All instruction halted until next key event)
		0xF00A: func(c *Chip8) {
			c.setVX(c.WaitKey())
			c.pc += 2
		},
		//FX15: Sets the delay timer to VX.
		0xF015: func(c *Chip8) {
			c.delayTimer = c.getVX()
			c.pc += 2
		},
		//FX18: Sets the sound timer to VX.
		0xF018: func(c *Chip8) {
			c.soundTimer = c.getVX()
			c.pc += 2
		},
		//FX1E: Adds VX to I.
		0xF01E: func(c *Chip8) {
			c.index += uint16(c.getVX())
			if c.index > 0xFFF {
				c.setVF(1)
				c.index &= 0xFFF
			} else {
				c.setVF(0)
			}
			c.pc += 2
		},
		//FX29: Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		0xF029: func(c *Chip8) {
			c.index = uint16(c.getVX() * 5)
			c.pc += 2
		},
		//FX33: Stores the binary-coded decimal representation of VX,
		//with the most significant of three digits at the address in I,
		//the middle digit at I plus 1, and the least significant digit at I plus 2.
		//(In other words, take the decimal representation of VX,
		//place the hundreds digit in memory at location in I, the tens digit at location I+1,
		//and the ones digit at location I+2.)
		0xF033: func(c *Chip8) {
			x := c.getVX()
			c.mem[c.index] = x / 100
			c.mem[c.index+1] = x % 100 / 10
			c.mem[c.index+2] = x % 10
			c.pc += 2
		},
		//FX55: Stores V0 to VX (including VX) in memory starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified.
		0xF055: func(c *Chip8) {
			for i := uint16(0); i <= (c.opcode&0x0F00)>>8; i++ {
				c.mem[c.index+i] = c.register[i]
			}
			c.pc += 2
		},
		//FX65: Fills V0 to VX (including VX) with values from memory starting at address I.
		//The offset from I is increased by 1 for each value written, but I itself is left unmodified.
		0xF065: func(c *Chip8) {
			for i := uint16(0); i <= (c.opcode&0x0F00)>>8; i++ {
				c.register[i] = c.mem[c.index+i]
			}
			c.pc += 2
		},
	}
)

type (
	moniter interface {
		Clear()
		Draw(int, int, []byte) byte
	}

	sounder interface {
		Beep()
	}

	inputer interface {
		IsPressed(byte) bool
		WaitKey() byte
	}

	//Chip8 is the atari vm
	Chip8 struct {
		mem        []byte
		opcode     uint16
		register   [16]byte
		index      uint16
		pc         uint16
		delayTimer byte
		soundTimer byte
		stack      [16]uint16
		sp         uint16

		codeKey   uint16
		cpuTick   <-chan time.Time
		timerTick <-chan time.Time

		moniter
		sounder
		inputer

		quit <-chan struct{}
	}
)

//Init the emulator
func (c *Chip8) Init(m moniter, s sounder, i inputer, quit <-chan struct{}) {
	c.moniter = m
	c.sounder = s
	c.inputer = i
	c.quit = quit

	c.cpuTick = time.Tick(cpuDuration)
	c.timerTick = time.Tick(timerDuration)

	c.pc = 0x200
	c.mem = make([]byte, 4096)
	copy(c.mem[:80], fontset)
	rand.Seed(time.Now().UnixNano())

	c.Clear()
}

//Load a game
func (c *Chip8) Load(r io.Reader) error {
	_, err := r.Read(c.mem[512:])
	return err
}

//Loop the game
func (c *Chip8) Loop() error {

	go c.countDown()

	var err error
loop:
	for {
		select {
		case <-c.cpuTick:
			c.fetch()

			c.decode()

			err = c.exec()
			if err != nil {
				break loop
			}
		case <-c.quit:
			break loop
		}
	}

	return err
}

func (c *Chip8) getVX() byte {
	return c.register[(c.opcode&0x0F00)>>8]
}
func (c *Chip8) setVX(b byte) {
	c.register[(c.opcode&0x0F00)>>8] = b
}

func (c *Chip8) getVY() byte {
	return c.register[(c.opcode&0x00F0)>>4]
}

func (c *Chip8) setVF(b byte) {
	c.register[0xF] = b
}

func (c *Chip8) getNNN() uint16 {
	return c.opcode & 0x0FFF
}

func (c *Chip8) getNN() byte {
	return byte(c.opcode & 0x00FF)
}

func (c *Chip8) fetch() {
	c.opcode = uint16(c.mem[c.pc])<<8 | uint16(c.mem[c.pc+1])
}

func (c *Chip8) decode() {
	switch c.opcode & 0xF000 {
	case 0x0000:
		c.codeKey = c.opcode
	case 0x8000:
		c.codeKey = c.opcode & 0xF00F
	case 0xE000, 0xF000:
		c.codeKey = c.opcode & 0xF0FF
	default:
		c.codeKey = c.opcode & 0xF000
	}
}

func (c *Chip8) exec() error {
	f, ok := funcmap[c.codeKey]
	if !ok {
		return fmt.Errorf("unknown opcode %X", c.opcode)
	}
	f(c)
	return nil
}

func (c *Chip8) countDown() {
	for range c.timerTick {
		if c.delayTimer > 0 {
			c.delayTimer--
		}
		if c.soundTimer > 0 {
			c.Beep()
			c.soundTimer--
		}
	}
}
