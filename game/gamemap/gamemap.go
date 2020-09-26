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
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/collision"
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/score"
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
	blockSpeed         = 25                               // our block speed
	boxSprite          = "box.png"                        // block sprite
	blockScale         = 0.5                              // block scale
	popSound           = "resources/audio/pop.wav"        // block pop
	hitSound           = "resources/audio/hit.wav"        // block hit
	font               = "resources/fonts/go_mono.fnt"    // our text font
	fontSize           = 30                               // top text fon size
	fontProduction     = "resources/fonts/go_regular.fnt" // our production text font
	fontProductionSize = 60                               // top production text fon size
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
	eng          *gosge.Engine     // the game engine
	length       int               // our map length
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
				ent := gms.sprs[c][r]
				ent.Remove(color.TYPE.Solid)
				ent.Remove(effects.TYPE.AlternateColor)
				ent.Remove(effects.TYPE.AlternateColorState)
				ent.Set(effects.AlternateColor{
					From:  color.Red,
					To:    color.Red.Alpha(127),
					Time:  0.25,
					Delay: 0,
				})
				pos := geometry.Get.Point(gms.sprs[c][r])
				if block.Text == nil {
					block.Text = gms.eng.World().AddEntity(
						ui.Text{
							String:     "0",
							Size:       fontSize,
							Font:       font,
							VAlignment: ui.MiddleVAlignment,
							HAlignment: ui.CenterHAlignment,
						},
						pos,
						color.White,
						movement.Movement{
							Amount: geometry.Point{
								X: -blockSpeed * gms.gs.Max,
							},
						},
						effects.Layer{Depth: -1},
					)
				}
				ent.Set(block)
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

	// pre-load font
	if err = eng.LoadFont(font); err != nil {
		return err
	}

	// pre-load production font
	if err = eng.LoadFont(fontProduction); err != nil {
		return err
	}

	// pop sound
	if err = eng.LoadSound(popSound); err != nil {
		return err
	}

	// hit sound
	if err = eng.LoadSound(hitSound); err != nil {
		return err
	}

	// get the block size
	if gms.blockSize, err = eng.GetSpriteSize(constants.SpriteSheet, boxSprite); err != nil {
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

	// we start on column 40
	cc := int((gms.dr.Width * gms.gs.Point.X) / (gms.blockSize.Width * blockScale * gms.gs.Max))

	// limits
	limitR := gms.rows - 8
	limitC := cc + gms.length

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
		cc += 20 + rand.Intn(2)
	}
}

// add sprite from map state
func (gms *gameMapSystem) addSprites(world *goecs.World) {
	offset := float32(0)

	// add a scroll marker
	gms.scrollMarker = gms.addEntity(world, 0, 0, offset)

	lastC := 0

	// for each column row
	for c := 0; c < gms.cols; c++ {
		for r := 0; r < gms.rows; r++ {
			// if empty skip
			if gms.data[c][r] == empty {
				continue
			}

			if c > lastC {
				lastC = c
			}

			// create a sprite
			ent := gms.addEntity(world, c, r, offset)

			ent.Add(sprite.Sprite{
				Sheet: constants.SpriteSheet,
				Name:  boxSprite,
				Scale: gms.gs.Max * blockScale,
			})

			clr := colors[(gms.data[c][r] - fill)]
			ent.Add(clr)
			ent.Add(effects.Layer{Depth: 0})
			ent.Add(component.Block{
				C: c,
				R: r,
			})
			gms.sprs[c][r] = ent
		}
	}

	textSize, _ := gms.eng.MeasureText(fontProduction, "Production", fontProductionSize)

	// add the production
	ent := gms.addEntity(world, lastC-10, gms.rows/2, offset)

	prodSize := geometry.Size{
		Width:  textSize.Width * 1.25 * gms.gs.Point.X,
		Height: gms.dr.Height * 0.90,
	}
	pos := geometry.Get.Point(ent)

	pos.Y = ((gms.dr.Height * gms.gs.Point.Y) - (prodSize.Height * gms.gs.Max)) * 0.5

	ent.Set(pos)
	ent.Add(shapes.SolidBox{
		Size:  prodSize,
		Scale: gms.gs.Max,
	})
	ent.Add(color.Gradient{
		From:      color.DarkBlue.Alpha(90),
		To:        color.SkyBlue.Alpha(70),
		Direction: color.GradientVertical,
	})
	ent.Add(effects.Layer{Depth: 1.0})

	ent = gms.addEntity(world, lastC+5, gms.rows/2, offset)

	ent.Set(pos)
	ent.Add(shapes.Box{
		Size:      prodSize,
		Scale:     gms.gs.Max,
		Thickness: int32(2 * gms.gs.Max),
	})
	ent.Add(color.White)
	ent.Add(effects.Layer{Depth: 1.0})

	ent = gms.addEntity(world, lastC+5, gms.rows/2, offset)
	pos.X += prodSize.Width * 0.5 * gms.gs.Max
	ent.Set(pos)
	ent.Add(ui.Text{
		String:     "Production",
		Size:       fontProductionSize * gms.gs.Max,
		Font:       fontProduction,
		VAlignment: ui.TopVAlignment,
		HAlignment: ui.CenterHAlignment,
	})
	ent.Add(color.White)
	ent.Add(effects.Layer{Depth: 1.0})
	ent.Add(component.Production{})
}

// create a entity on that col and row, with movement
func (gms *gameMapSystem) addEntity(world *goecs.World, col, row int, offset float32) *goecs.Entity {
	px := float32(col) * (gms.blockSize.Width * gms.gs.Max * blockScale)
	px += (gms.blockSize.Width / 2) * gms.gs.Max * blockScale
	px += offset
	py := float32(row) * (gms.blockSize.Height * gms.gs.Max * blockScale)
	py += (gms.blockSize.Height / 2) * gms.gs.Max * blockScale

	return world.AddEntity(
		geometry.Point{
			X: px,
			Y: py,
		},
		movement.Movement{
			Amount: geometry.Point{
				X: -blockSpeed * gms.gs.Max,
			},
		},
	)
}

func (gms *gameMapSystem) bulletSystem(world *goecs.World, _ float32) error {
	for it := world.Iterator(component.TYPE.Bullet, geometry.TYPE.Point); it != nil; it = it.Next() {
		bullet := it.Value()
		pos := geometry.Get.Point(bullet)
		if pos.X >= gms.dr.Width*gms.gs.Max {
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
			x := pos.X - gms.blockSize.Width*0.5*blockScale*gms.gs.Max

			// create a sprite
			nb := gms.addEntity(world, c, r, x)

			nb.Add(sprite.Sprite{
				Sheet: constants.SpriteSheet,
				Name:  boxSprite,
				Scale: gms.gs.Max * blockScale,
			})

			nb.Add(color.Red)

			nb.Add(effects.Layer{Depth: 0})
			nb.Add(component.Block{
				C: c,
				R: r,
			})
			gms.sprs[c][r] = nb
			gms.place(c, r)
			return world.Signal(events.PlaySoundEvent{Name: hitSound, Volume: 1})
		}
	case collision.PlaneHitBlockEvent:
		block := e.Block
		at := geometry.Get.Point(gms.sprs[block.C][block.R])
		if err := world.Signal(score.PointsEvent{Total: -1, At: at}); err != nil {
			return err
		}
		gms.data[block.C][block.R] = clear
		gms.sprs[block.C][block.R] = nil
		return world.Signal(events.PlaySoundEvent{Name: hitSound, Volume: 1})
	case collision.MeshHitBlockEvent:
		block := e.Block
		at := geometry.Get.Point(gms.sprs[block.C][block.R])
		if err := world.Signal(score.PointsEvent{Total: -2, At: at}); err != nil {
			return err
		}
		gms.data[block.C][block.R] = clear
		gms.sprs[block.C][block.R] = nil
		return world.Signal(events.PlaySoundEvent{Name: hitSound, Volume: 1})
	}

	return nil
}

func (gms *gameMapSystem) clearSystem(world *goecs.World, delta float32) error {
	// total block we clear
	total := 0

	// total x and y for the block that we clear
	totalX := float32(0)
	totalY := float32(0)

	// iterate the blocks
	for it := world.Iterator(component.TYPE.Block); it != nil; it = it.Next() {
		ent := it.Value()
		block := component.Get.Block(ent)
		// if is a block that need clear
		if gms.data[block.C][block.R] == clear {
			// decrease time
			block.ClearOn -= delta
			// update block
			ent.Set(block)
			// if we are on time to clear
			if block.ClearOn <= 0 {
				// get the position and add to the totals
				pos := geometry.Get.Point(ent)
				totalX += pos.X
				totalY += pos.Y
				// remove text
				_ = world.Remove(block.Text)
				block.Text = nil
				ent.Set(block)
				// remove entity
				_ = world.Remove(ent)
				// remove from our slices
				gms.data[block.C][block.R] = empty
				gms.sprs[block.C][block.R] = nil
				total++
			} else {
				sec := fmt.Sprintf("%0.0f", block.ClearOn)
				text := ui.Get.Text(block.Text)
				text.String = sec
				block.Text.Set(text)
			}
		}
	}
	// if we have clear any block
	if total > 0 {
		var err error
		// the points are generate at the average of all blocks position
		at := geometry.Point{
			X: totalX / float32(total),
			Y: totalY / float32(total),
		}
		// signal that we got points at a position
		if err = world.Signal(score.PointsEvent{Total: total, At: at}); err != nil {
			return err
		}

		// play pop sound
		if err = world.Signal(events.PlaySoundEvent{Name: popSound, Volume: 1}); err != nil {
			return err
		}
	}

	return nil
}

// System create the map system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size, length int) error {
	gms := newGameMap(length+100, 34)

	gms.length = length
	gms.gs = gs
	gms.dr = dr
	gms.eng = engine

	return gms.load(engine)
}
