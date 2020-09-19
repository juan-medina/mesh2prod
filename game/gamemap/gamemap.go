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
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/mesh2prod/game/collision"
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type blocState int

//goland:noinspection GoUnusedConst
const (
	empty = blocState(iota)
	placed
	clear
	fill
)

// logic constants
const (
	blockSpeed      = 25              // our block speed
	containerSprite = "container.png" // block sprite
	sidecarSprite   = "sidecar.png"   // block sprite
	blockScale      = 0.5             // block scale
)

type gameMapSystem struct {
	rows         int               // number of rows
	cols         int               // number of cols
	data         [][]blocState     // map block state
	sprs         [][]*goecs.Entity // map sprites
	gs           geometry.Scale    // game scale
	dr           geometry.Size     // design resolution
	blockSize    geometry.Size     // block size
	scrollMarker *goecs.Entity     // track the scroll position
}

var (
	colors = []color.Solid{
		color.Yellow,
		color.Gold,
		color.Orange,
		color.Pink,
		color.Green,
		color.Purple,
		color.Beige,
		color.Gopher,
	}
)

// generate a string for current map status
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

// set the status on map block
func (gms *gameMapSystem) set(c, r int, state blocState) {
	gms.data[c][r] = state
}

// place a block and mark the block that need to be clear
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

	// check for areas
	for cc := c + 1; cc <= sc; cc++ {
		// areas on top of the place block
		for cr := r - 1; cr >= tr; cr-- {
			can := gms.canClearArea(c, cr, cc, r)
			if can {
				gms.clearArea(c, cr, cc, r)
			}
		}
		// areas under the place block
		for cr := br; cr > r; cr-- {
			can := gms.canClearArea(c, r, cc, cr)
			if can {
				gms.clearArea(c, r, cc, cr)
			}
		}
	}
}

// add a block in a position
func (gms *gameMapSystem) add(col, row int, piece [][]blocState, color int) {
	for r := 0; r < len(piece); r++ {
		for c := 0; c < len(piece[r]); c++ {
			if piece[r][c] != empty {
				gms.data[col+c][row+r] = blocState(int(piece[r][c]) + color)
			}
		}
	}
}

// can we clear an area? is a square area
func (gms *gameMapSystem) canClearArea(fromC, fromR, toC, toR int) bool {
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

// clear the area
func (gms *gameMapSystem) clearArea(fromC, fromR, toC, toR int) {
	for c := fromC; c <= toC; c++ {
		for r := fromR; r <= toR; r++ {
			gms.data[c][r] = clear
			if gms.sprs[c][r] != nil {
				block := component.Get.Block(gms.sprs[c][r])
				block.ClearOn = 5
				gms.sprs[c][r].Set(block)
				gms.sprs[c][r].Remove(color.TYPE.Solid)
				gms.sprs[c][r].Remove(effects.TYPE.AlternateColor)
				gms.sprs[c][r].Remove(effects.TYPE.AlternateColorState)
				gms.sprs[c][r].Set(effects.AlternateColor{
					From:  color.Red,
					To:    color.Red.Blend(color.Black, 0.50).Alpha(127),
					Time:  0.25,
					Delay: 0,
				})
			}
		}
	}
}

// create a new game map
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

// create a map from an string
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

// load the system
func (gms *gameMapSystem) load(eng *gosge.Engine) error {
	rand.Seed(time.Now().UnixNano())
	var err error

	// get the block size
	if gms.blockSize, err = eng.GetSpriteSize(constants.SpriteSheet, containerSprite); err != nil {
		return err
	}

	// generate a random map
	gms.generate()

	// get the world
	world := eng.World()

	// add the sprites from the current state
	gms.addSprites(world)

	// add the bullet system
	world.AddSystem(gms.bulletSystem)

	// clear block systems
	world.AddSystem(gms.clearSystem)

	// listen to collisions
	world.AddListener(gms.collisionListener)

	return nil
}

// generate a random map
func (gms *gameMapSystem) generate() {
	// pieces
	piece1 := [][]blocState{
		{0, 3},
		{3, 3},
		{3, 3},
		{0, 3},
	}

	piece2 := [][]blocState{
		{3, 3, 3},
		{0, 3, 3},
		{3, 3, 3},
	}

	piece3 := [][]blocState{
		{3, 3, 3},
		{0, 3, 3},
		{0, 3, 3},
	}

	piece4 := [][]blocState{
		{0, 3, 3},
		{0, 3, 3},
		{3, 3, 3},
	}

	piece5 := [][]blocState{
		{0, 3},
		{3, 3},
	}

	piece6 := [][]blocState{
		{3, 3},
		{0, 3},
	}

	piece7 := [][]blocState{
		{3, 3},
		{0, 3},
		{0, 3},
		{3, 3},
	}

	piece8 := [][]blocState{
		{3, 3, 3, 3},
		{0, 3, 3, 3},
		{0, 0, 3, 3},
		{0, 0, 0, 3},
	}

	piece9 := [][]blocState{
		{0, 0, 0, 3},
		{0, 0, 3, 3},
		{0, 3, 3, 3},
		{3, 3, 3, 3},
	}

	// set o pieces
	pieces := [][][]blocState{
		piece1,
		piece2,
		piece3,
		piece4,
		piece5,
		piece6,
		piece7,
		piece8,
		piece9,
	}

	// limits
	limitR := gms.rows - 8
	limitC := gms.cols - 8

	// we start on column 4
	cc := 6
	for cc < limitC {
		// random number of pieces
		num := 2 + rand.Intn(6)
		// fil the pieces
		for i := 0; i < num; i++ {
			// random shift of column
			c := cc - rand.Intn(6)
			// random piece
			p := rand.Intn(len(pieces))
			// random shift of row
			r := 4 + rand.Intn(limitR)
			clr := rand.Intn(len(colors))
			// add piece
			gms.add(c, r, pieces[p], clr)
		}

		// advance column random
		cc += 15 + rand.Intn(5)
	}
}

// add sprite from map state
func (gms *gameMapSystem) addSprites(world *goecs.World) {
	offset := gms.dr.Width * .85 * gms.gs.Point.X

	// add a scroll marker
	gms.scrollMarker = gms.addEntity(world, 0, 0, offset)

	// for each column row
	for c := 0; c < gms.cols; c++ {
		for r := 0; r < gms.rows; r++ {
			// if empty skip
			if gms.data[c][r] == empty {
				continue
			}

			// create a sprite
			ent := gms.addEntity(world, c, r, offset)

			ent.Add(sprite.Sprite{
				Sheet: constants.SpriteSheet,
				Name:  containerSprite,
				Scale: gms.gs.Min * blockScale,
			})

			clr := colors[(gms.data[c][r] - fill)]

			ent.Add(effects.AlternateColor{
				From:  clr,
				To:    clr.Blend(color.Black, 0.15),
				Time:  0.25,
				Delay: 0,
			})

			ent.Add(effects.Layer{Depth: 0})
			ent.Add(component.Block{
				C: c,
				R: r,
			})
			gms.sprs[c][r] = ent
		}
	}
}

// create a entity on that col and row, with movement
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

func (gms *gameMapSystem) bulletSystem(world *goecs.World, _ float32) error {
	for it := world.Iterator(component.TYPE.Bullet, geometry.TYPE.Point); it != nil; it = it.Next() {
		bullet := it.Value()
		pos := geometry.Get.Point(bullet)
		if pos.X >= gms.dr.Width*gms.gs.Point.X {
			_ = world.Remove(bullet)
		}
	}
	return nil
}

func (gms *gameMapSystem) collisionListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case collision.BulletHitBlockEvent:
		// get the current scroll
		pos := geometry.Get.Point(gms.scrollMarker)
		c := e.Block.C - 1
		r := e.Block.R
		if c > 0 {
			x := pos.X - gms.blockSize.Width*0.5*blockScale*gms.gs.Point.X

			// create a sprite
			nb := gms.addEntity(world, c, r, x)

			nb.Add(sprite.Sprite{
				Sheet: constants.SpriteSheet,
				Name:  sidecarSprite,
				Scale: gms.gs.Min * blockScale,
			})

			nb.Add(color.Red)

			nb.Add(effects.Layer{Depth: 0})
			nb.Add(component.Block{
				C: c,
				R: r,
			})
			gms.sprs[c][r] = nb
			gms.place(c, r)
		}

	}

	return nil
}

func (gms *gameMapSystem) clearSystem(world *goecs.World, delta float32) error {
	for it := world.Iterator(component.TYPE.Block); it != nil; it = it.Next() {
		ent := it.Value()
		block := component.Get.Block(ent)
		if gms.data[block.C][block.R] == clear {
			block.ClearOn -= delta
			ent.Set(block)
			if block.ClearOn <= 0 {
				_ = world.Remove(ent)
				gms.data[block.C][block.R] = empty
				gms.sprs[block.C][block.R] = nil
			}
		}
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
