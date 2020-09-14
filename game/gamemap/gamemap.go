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

func (g *gameMap) Set(c, r int, state blocState) {
	g.data[c][r] = state
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
