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
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/plane"
	"math/rand"
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

// logic constants
const (
	blockSpeed  = 50
	blockSprite = "block.png"
	markSprite  = "mark.png"
	blockScale  = 0.5
	targetGapX  = 100
)

type gameMapSystem struct {
	rows         int
	cols         int
	data         [][]blocState
	sprs         [][]*goecs.Entity
	gs           geometry.Scale
	dr           geometry.Size
	blockSize    geometry.Size
	scrollMarker *goecs.Entity
	gunPos       geometry.Point
	cursor       *goecs.Entity
	line         *goecs.Entity
}

func (gms gameMapSystem) String() string {
	var result = ""
	for r := 0; r < gms.rows; r++ {
		for c := 0; c < gms.cols; c++ {
			block := gms.data[c][r]
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

func (gms *gameMapSystem) set(c, r int, state blocState) {
	gms.data[c][r] = state
}

func (gms *gameMapSystem) place(c, r int) {
	// we set this block to place
	gms.data[c][r] = placed

	// search the top row
	var tr int
	for tr = r; tr >= 0; tr-- {
		if gms.data[c][tr] == empty {
			tr++
			break
		}
	}

	// search the right column
	var sc int
	for sc = c; sc < gms.cols; sc++ {
		if gms.data[sc][r] == empty {
			sc--
			break
		}
	}

	// search the bottom row
	var br int
	for br = r; br < gms.rows; br++ {
		if gms.data[c][br] == empty {
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
			if gms.canClearArea(fromC, fromR, cc, cr) {
				gms.clearArea(fromC, fromR, cc, cr)
			}
		}
	}
}

func (gms *gameMapSystem) add(col, row int, piece [][]blocState) {
	for r := 0; r < len(piece); r++ {
		for c := 0; c < len(piece[r]); c++ {
			if piece[r][c] != empty {
				gms.data[col+c][row+r] = piece[r][c]
			}
		}
	}
}

func (gms gameMapSystem) canClearArea(fromC, fromR, toC, toR int) bool {
	// top row
	for c := fromC; c <= toC; c++ {
		if gms.data[c][fromR] == empty {
			return false
		}
	}

	// bottom row
	for c := fromC; c <= toC; c++ {
		if gms.data[c][toR] == empty {
			return false
		}
	}

	// left column
	for r := fromR; r <= toR; r++ {
		if gms.data[fromC][r] == empty {
			return false
		}
	}

	// right column
	for r := fromR; r <= toR; r++ {
		if gms.data[toC][r] == empty {
			return false
		}
	}

	return true
}

func (gms *gameMapSystem) clearArea(fromC, fromR, toC, toR int) {
	for c := fromC; c <= toC; c++ {
		for r := fromR; r <= toR; r++ {
			gms.data[c][r] = clear
		}
	}
}

func newGameMap(cols, rows int) *gameMapSystem {
	data := make([][]blocState, cols)
	sprs := make([][]*goecs.Entity, cols)
	for c := 0; c < cols; c++ {
		data[c] = make([]blocState, rows)
		sprs[c] = make([]*goecs.Entity, rows)
	}
	return &gameMapSystem{
		rows: rows,
		cols: cols,
		data: data,
		sprs: sprs,
	}
}

func fromString(str string) *gameMapSystem {
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

func (gms *gameMapSystem) load(eng *gosge.Engine) error {
	var err error

	if gms.blockSize, err = eng.GetSpriteSize(constants.SpriteSheet, blockSprite); err != nil {
		return err
	}

	gms.generate()

	world := eng.World()

	gms.addSprites(world)
	world.AddSystem(gms.cursorSystem)
	world.AddListener(gms.planeChanges)

	return nil
}

func (gms *gameMapSystem) generate() {

	piece1 := [][]blocState{
		{0, 1},
		{1, 1},
		{1, 1},
		{0, 1},
	}

	piece2 := [][]blocState{
		{1, 1, 1},
		{0, 1, 1},
		{1, 1, 1},
	}

	piece3 := [][]blocState{
		{1, 1, 1},
		{0, 1, 1},
		{0, 1, 1},
	}

	piece4 := [][]blocState{
		{0, 1, 1},
		{0, 1, 1},
		{1, 1, 1},
	}

	piece5 := [][]blocState{
		{0, 1},
		{1, 1},
	}

	piece6 := [][]blocState{
		{1, 1},
		{0, 1},
	}

	piece7 := [][]blocState{
		{1, 1},
		{0, 1},
		{0, 1},
		{1, 1},
	}

	pieces := [][][]blocState{
		piece1,
		piece2,
		piece3,
		piece4,
		piece5,
		piece6,
		piece7,
	}

	limitR := gms.rows - 8
	limitC := gms.cols - 8

	cc := 4
	for cc < limitC {
		num := 2 + rand.Intn(6)
		for i := 0; i < num; i++ {
			c := cc - rand.Intn(6)
			p := rand.Intn(len(pieces))
			r := 4 + rand.Intn(limitR)
			gms.add(c, r, pieces[p])
		}

		cc += 15 + rand.Intn(5)
	}

}

func (gms *gameMapSystem) addSprites(world *goecs.World) {
	offset := gms.dr.Width * .85 * gms.gs.Point.X

	gms.scrollMarker = gms.addEntity(world, 0, 0, offset)

	gms.cursor = gms.addEntity(world, 0, 0, offset)

	gms.cursor.Add(sprite.Sprite{
		Sheet: constants.SpriteSheet,
		Name:  markSprite,
		Scale: gms.gs.Min * blockScale,
	})
	gms.cursor.Add(effects.Layer{Depth: 0})
	gms.cursor.Add(color.Red)
	gms.cursor.Add(effects.AlternateColor{
		From:  color.Red,
		To:    color.Red.Alpha(180),
		Time:  0.25,
		Delay: 0,
	})

	gms.line = gms.addEntity(world, 0, 0, offset)
	gms.line.Add(color.Red.Alpha(127))
	gms.line.Add(shapes.Line{
		To:        geometry.Point{},
		Thickness: 2,
	})
	gms.line.Add(effects.Layer{Depth: 0})

	for c := 0; c < gms.cols; c++ {
		for r := 0; r < gms.rows; r++ {
			if gms.data[c][r] == empty {
				continue
			}

			ent := gms.addEntity(world, c, r, offset)

			ent.Add(sprite.Sprite{
				Sheet: constants.SpriteSheet,
				Name:  blockSprite,
				Scale: gms.gs.Min * blockScale,
			})

			ent.Add(effects.AlternateColor{
				From:  color.Gopher,
				To:    color.SkyBlue,
				Time:  0.25,
				Delay: 0,
			})

			ent.Add(effects.Layer{Depth: 0})
			gms.sprs[c][r] = ent
		}
	}
}

func (gms *gameMapSystem) addEntity(world *goecs.World, col, row int, offset float32) *goecs.Entity {
	px := float32(col) * (gms.blockSize.Width * gms.gs.Point.X * blockScale)
	px += (gms.blockSize.Width / 2) * gms.gs.Point.X * blockScale
	px += offset
	py := float32(row) * (gms.blockSize.Height * gms.gs.Point.Y * blockScale)
	py += (gms.blockSize.Height / 2) * gms.gs.Point.Y * blockScale

	return world.AddEntity(
		geometry.Point{
			X: px,
			Y: py,
		},
		movement.Movement{
			Amount: geometry.Point{
				X: -blockSpeed * gms.gs.Point.X,
			},
		},
	)
}

func (gms *gameMapSystem) cursorSystem(_ *goecs.World, _ float32) error {
	pos := geometry.Get.Point(gms.scrollMarker)
	x := gms.gunPos.X - pos.X
	y := gms.gunPos.Y - pos.Y
	c := int(x / (gms.blockSize.Width * gms.gs.Point.X * blockScale))
	r := int(y / (gms.blockSize.Height * gms.gs.Point.Y * blockScale))

	linePosFrom := geometry.Get.Point(gms.line)
	linePosFrom.X = gms.gunPos.X
	linePosFrom.Y = gms.gunPos.Y
	line := shapes.Get.Line(gms.line)

	found := false
	var sc = c
	for sc = c; sc < gms.cols; sc++ {
		if sc >= 0 {
			if gms.data[sc][r] != empty {
				pos := geometry.Get.Point(gms.sprs[sc][r])
				if pos.X > (gms.gunPos.X+(targetGapX*gms.gs.Point.X)) &&
					(pos.X < (gms.dr.Width * gms.gs.Point.X)) {
					found = true
					cursorPos := geometry.Point{
						X: pos.X - (gms.blockSize.Width * gms.gs.Point.X * blockScale),
						Y: pos.Y,
					}
					gms.cursor.Set(cursorPos)
					line.To = geometry.Point{
						X: cursorPos.X - (gms.blockSize.Width/2)*blockScale*gms.gs.Point.X,
						Y: cursorPos.Y,
					}
					//linePosFrom.Y = cursorPos.Y
				}
				break
			}
		}
	}

	if !found {
		gms.cursor.Set(geometry.Point{
			X: -1000,
			Y: -1000,
		})

		line.To = geometry.Point{
			X: gms.dr.Width * gms.gs.Point.X,
			Y: gms.gunPos.Y,
		}
	}

	gms.line.Set(linePosFrom)
	gms.line.Set(line)

	return nil
}

func (gms *gameMapSystem) planeChanges(_ *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		gms.gunPos = e.Gun
	}
	return nil
}

// System create the map system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	gms := newGameMap(500, 34)

	gms.gs = gs
	gms.dr = dr

	return gms.load(engine)
}
