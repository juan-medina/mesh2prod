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
	"github.com/juan-medina/gosge/components/animation"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/collision"
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/plane"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
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
	blockSpeed        = 25              // our block speed
	blockSprite       = "block.png"     // block sprite
	markSprite        = "mark.png"      // mark sprite
	blockScale        = 0.5             // block scale
	targetGapX        = 100             // target gap from gun pos
	bulletSprite      = "bullet_%d.png" // bullet sprite base
	bulletScale       = 0.25            // scale for the bullet sprite
	bulletFrames      = 5               // bullet frames
	bulletFramesDelay = 0.065           // bullet frame delay
	bulletSpeed       = 600             // bullet speed
)

var (
	bulletColor = color.Red.Alpha(180) // bullet color
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
	gunPos       geometry.Point    // plane gun position
	target       *goecs.Entity     // current target position
	line         *goecs.Entity     // target line
	lastShoot    float32
}

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
func (gms *gameMapSystem) add(col, row int, piece [][]blocState) {
	for r := 0; r < len(piece); r++ {
		for c := 0; c < len(piece[r]); c++ {
			if piece[r][c] != empty {
				gms.data[col+c][row+r] = piece[r][c]
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
				gms.sprs[c][r].Set(block)
				gms.sprs[c][r].Remove(color.TYPE.Solid)
				gms.sprs[c][r].Remove(effects.TYPE.AlternateColor)
				gms.sprs[c][r].Remove(effects.TYPE.AlternateColorState)
				gms.sprs[c][r].Set(effects.AlternateColor{
					From:  color.Red,
					To:    color.SkyBlue,
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
	if gms.blockSize, err = eng.GetSpriteSize(constants.SpriteSheet, blockSprite); err != nil {
		return err
	}

	// generate a random map
	gms.generate()

	// get the world
	world := eng.World()

	// add the sprites from the current state
	gms.addSprites(world)

	// add the target system that target blocks
	world.AddSystem(gms.targetSystem)

	// add the bullet system
	world.AddSystem(gms.bulletSystem)

	// clear block systems
	world.AddSystem(gms.clearSystem)

	// listen to plane changes
	world.AddListener(gms.planeChanges)

	// listen to keys
	world.AddListener(gms.keyListener)

	// listen to collisions
	world.AddListener(gms.collisionListener)

	return nil
}

// generate a random map
func (gms *gameMapSystem) generate() {
	// pieces
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

	piece8 := [][]blocState{
		{1, 1, 1, 1},
		{0, 1, 1, 1},
		{0, 0, 1, 1},
		{0, 0, 0, 1},
	}

	piece9 := [][]blocState{
		{0, 0, 0, 1},
		{0, 0, 1, 1},
		{0, 1, 1, 1},
		{1, 1, 1, 1},
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
			// add piece
			gms.add(c, r, pieces[p])
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

	// add our target
	gms.target = gms.addEntity(world, 0, 0, offset)

	gms.target.Add(sprite.Sprite{
		Sheet: constants.SpriteSheet,
		Name:  markSprite,
		Scale: gms.gs.Min * blockScale,
	})
	gms.target.Add(effects.Layer{Depth: 0})
	gms.target.Add(color.Red)
	gms.target.Add(effects.AlternateColor{
		From:  color.Red,
		To:    color.Red.Alpha(180),
		Time:  0.25,
		Delay: 0,
	})

	// add the target line
	gms.line = gms.addEntity(world, 0, 0, offset)
	gms.line.Add(color.Red.Alpha(127))
	gms.line.Add(shapes.Line{
		To:        geometry.Point{},
		Thickness: 2,
	})
	gms.line.Add(effects.Layer{Depth: 0})

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

// a system that target a block
func (gms *gameMapSystem) targetSystem(_ *goecs.World, _ float32) error {
	// get the current scroll
	pos := geometry.Get.Point(gms.scrollMarker)
	// get displacement for our gun
	x := gms.gunPos.X - (pos.X - (gms.blockSize.Width / 2))
	y := gms.gunPos.Y - (pos.Y - (gms.blockSize.Height / 2))
	// calculate row and column
	c := int(x / (gms.blockSize.Width * blockScale * gms.gs.Point.X))
	r := int(y / (gms.blockSize.Height * blockScale * gms.gs.Point.Y))

	if c < 0 {
		c = 0
	}

	if r < 0 {
		r = 0
	}

	// get the line from
	linePosFrom := geometry.Get.Point(gms.line)
	linePosFrom.X = gms.gunPos.X
	linePosFrom.Y = gms.gunPos.Y
	// get the line component
	line := shapes.Get.Line(gms.line)

	// try to find a target
	found := false
	var sc = c
	// goes trough the rows
	for sc = c; sc < gms.cols; sc++ {
		if sc >= 0 {
			// if the block is no empty
			if gms.data[sc][r] != empty && gms.sprs[sc][r] != nil {
				pos := geometry.Get.Point(gms.sprs[sc][r])
				// if we are withing the gap and the screen size
				if pos.X > (gms.gunPos.X+(targetGapX*gms.gs.Point.X)) &&
					(pos.X < (gms.dr.Width * gms.gs.Point.X)) {
					// found it
					found = true
					// calculate target pos
					targetPos := geometry.Point{
						X: pos.X - (gms.blockSize.Width * gms.gs.Point.X * blockScale),
						Y: pos.Y,
					}
					gms.target.Set(targetPos)
					// calculate line pos
					line.To = geometry.Point{
						X: targetPos.X - (gms.blockSize.Width/2)*blockScale*gms.gs.Point.X,
						Y: targetPos.Y,
					}
				}
				// it was a block no empty, skip
				break
			}
		}
	}

	// if we have no a target
	if !found {
		// move target ouf ot screen
		gms.target.Set(geometry.Point{
			X: -1000,
			Y: -1000,
		})

		// move line straight from gun
		line.To = geometry.Point{
			X: gms.dr.Width * gms.gs.Point.X,
			Y: gms.gunPos.Y,
		}
	}

	// update line
	gms.line.Set(linePosFrom)
	gms.line.Set(line)

	return nil
}

// if the plane change position
func (gms *gameMapSystem) planeChanges(_ *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		// store gun position
		gms.gunPos = e.Gun
	}
	return nil
}

// listen to keys
func (gms *gameMapSystem) keyListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	// if we got a key up
	case events.KeyUpEvent:
		// if it space
		if e.Key == device.KeySpace {
			gms.createBullet(world)
			gms.lastShoot = 0
		}
	}
	return nil
}

func (gms gameMapSystem) createBullet(world *goecs.World) {
	// get target
	targetPos := geometry.Get.Point(gms.target)
	// if we have a target on the screen
	if targetPos.X > 0 && targetPos.Y > 0 {
		// calculate min / max y and velocity
		minY := gms.gunPos.Y
		maxY := targetPos.Y
		velY := maxY - minY
		velY = float32(float64(velY)/math.Abs(float64(velY))) * bulletSpeed * 10
		if minY > maxY {
			aux := minY
			minY = maxY
			maxY = aux
		}
		// add a bullet
		world.AddEntity(
			animation.Animation{
				Sequences: map[string]animation.Sequence{
					"moving": {
						Sheet:  constants.SpriteSheet,
						Base:   bulletSprite,
						Scale:  gms.gs.Min * bulletScale,
						Frames: bulletFrames,
						Delay:  bulletFramesDelay,
					},
				},
				Current: "moving",
				Speed:   1,
			},
			gms.gunPos,
			movement.Movement{
				Amount: geometry.Point{
					Y: velY * gms.gs.Point.Y,
					X: bulletSpeed * gms.gs.Point.X,
				},
			},
			movement.Constrain{
				Min: geometry.Point{
					X: 0,
					Y: minY,
				},
				Max: geometry.Point{
					X: gms.dr.Width * gms.gs.Point.X,
					Y: maxY,
				},
			},
			bulletColor,
			component.Bullet{},
			effects.Layer{Depth: 0},
		)
	}
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
				Name:  blockSprite,
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
	gms.lastShoot += delta
	if gms.lastShoot > 2 {
		for it := world.Iterator(component.TYPE.Block); it != nil; it = it.Next() {
			ent := it.Value()
			block := component.Get.Block(ent)
			if gms.data[block.C][block.R] == clear {
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
