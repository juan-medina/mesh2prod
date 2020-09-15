/*
 * Copyright (c) 2020 Juan Medina.
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in
 *  all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *  THE SOFTWARE.
 */

package gamemap

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type blocState int

//goland:noinspection GoUnusedConst
const (
	empty = blocState(iota)
	fill
	placed
	clear
)

type gameMap struct {
	rows int
	cols int
	data [][]blocState
}

func (g gameMap) String() string {
	var result = ""
	for r := 0; r < g.rows; r++ {
		for c := 0; c < g.cols; c++ {
			block := g.data[c][r]
			if block != empty {
				result += strconv.Itoa(int(block))
			} else {
				result += " "
			}

		}
		result += "\n"
	}
	return result
}

func (g *gameMap) set(c, r int, state blocState) {
	g.data[c][r] = state
}

func (g *gameMap) place(c, r int) {
	// we set this block to place
	g.data[c][r] = placed

	// search the top row
	var tr int
	for tr = r; tr >= 0; tr-- {
		if g.data[c][tr] == empty {
			tr++
			break
		}
	}

	// search the right column
	var sc int
	for sc = c; sc < g.cols; sc++ {
		if g.data[sc][r] == empty {
			sc--
			break
		}
	}

	// search the bottom row
	var br int
	for br = r; br < g.rows; br++ {
		if g.data[c][br] == empty {
			br--
			break
		}
	}

	// calculate from and to
	var fromC, fromR, toC, toR int
	fromC = c
	fromR = tr
	toC = sc
	toR = br

	// check for the whole area each subarea
	for cc := fromC + 1; cc <= toC; cc++ {
		for cr := fromR + 1; cr <= toR; cr++ {
			// if we could clear this sub area clear it
			if g.canClearArea(fromC, fromR, cc, cr) {
				g.clearArea(fromC, fromR, cc, cr)
			}
		}
	}
}

func (g *gameMap) add(col, row int, piece [][]blocState) {
	for r := 0; r < len(piece); r++ {
		for c := 0; c < len(piece[r]); c++ {
			g.data[col+c][row+r] = piece[r][c]
		}
	}
}

func (g gameMap) canClearArea(fromC, fromR, toC, toR int) bool {
	for c := fromC; c <= toC; c++ {
		if g.data[c][fromR] == empty {
			return false
		}
	}

	for c := fromC; c <= toC; c++ {
		if g.data[c][toR] == empty {
			return false
		}
	}

	for r := fromR; r <= toR; r++ {
		if g.data[fromC][r] == empty {
			return false
		}
	}

	for r := fromR; r <= toR; r++ {
		if g.data[toC][r] == empty {
			return false
		}
	}

	return true
}

func (g *gameMap) clearArea(fromC, fromR, toC, toR int) {
	for c := fromC; c <= toC; c++ {
		for r := fromR; r <= toR; r++ {
			g.data[c][r] = clear
		}
	}
}

func newGameMap(cols, rows int) *gameMap {
	data := make([][]blocState, cols)
	for c := 0; c < cols; c++ {
		data[c] = make([]blocState, rows)
	}
	return &gameMap{
		rows: rows,
		cols: cols,
		data: data,
	}
}

func fromString(str string) *gameMap {
	scanner := bufio.NewScanner(strings.NewReader(str))
	r := 0
	c := 0
	for scanner.Scan() {
		s := len(scanner.Text())
		if s > c {
			c = s
		}
		r++
	}

	gm := newGameMap(c, r)
	scanner = bufio.NewScanner(strings.NewReader(str))

	r = 0
	for scanner.Scan() {
		t := scanner.Text()
		c = 0
		for _, d := range t {
			b, _ := strconv.Atoi(fmt.Sprintf("%c", d))
			gm.data[c][r] = blocState(b)
			c++
		}
		r++
	}

	return gm
}
