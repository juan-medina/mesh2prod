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

package score

import (
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/winning"
	"math"
)

// PointsEvent is trigger when new points need to be added
type PointsEvent struct {
	Total int
	At    geometry.Point
}

// logic constants
const (
	font                   = "resources/fonts/go_mono.fnt" // our text font
	fontSize               = 40                            // top text fon size
	floatPointSize         = 30                            // float text font size
	blockChainSprite       = "blockchain.png"              // blockchain coin
	blockChainScale        = 0.5                           // blockchain coin scale
	bcGapY                 = 5                             // blockchain coin gap Y
	bcGapX                 = 5                             // blockchain coin gap X
	pointPerBlock          = 5                             // points given per each block
	pointLosePerBlock      = 20                            // points loosing when hit the plane
	textAlphaDecreesPerSec = 127                           // alpha decrease per sec in floating text
	textScrollSpeedY       = 100                           // text scroll y
	textScrollSpeedX       = 25                            // text scroll x (match block scroll)
	pointsToAddPerSec      = 100                           // points to add each second
)

type scoreSystem struct {
	gs        geometry.Scale // game scale
	dr        geometry.Size  // design resolution
	total     int            // total points
	lastScore int            // last score
	toAdd     int            // score to add
	toSub     int            // score to sub
	textLabel *goecs.Entity  // our text
	end       bool
}

var (
	bcColor       = color.Solid{R: 227, G: 140, B: 41, A: 255} // our bc text color
	positiveColor = color.Green                                // positive numbers color
	negativeColor = color.Red                                  // negative numbers color
)

// load the system
func (ss *scoreSystem) load(eng *gosge.Engine) error {
	var err error

	// get the world
	world := eng.World()

	// pre-load font
	if err = eng.LoadFont(font); err != nil {
		return err
	}

	var size geometry.Size

	// get the bc coin size
	if size, err = eng.GetSpriteSize(constants.SpriteSheet, blockChainSprite); err != nil {
		return err
	}

	// calculate the bc coin position, top right
	bcPos := geometry.Point{
		X: (ss.dr.Width * ss.gs.Point.X) - (size.Width * blockChainScale * 0.5 * ss.gs.Max),
		Y: size.Height * 0.5 * blockChainScale * ss.gs.Point.Y,
	}

	// add a extra gap
	bcPos.X -= bcGapX * ss.gs.Max
	bcPos.Y += bcGapY * ss.gs.Max

	// add the bc coin
	world.AddEntity(
		sprite.Sprite{
			Sheet: constants.SpriteSheet,
			Name:  blockChainSprite,
			Scale: ss.gs.Max * blockChainScale,
		},
		bcPos,
		effects.Layer{Depth: -10},
	)

	// add the text label
	ss.textLabel = world.AddEntity(
		ui.Text{
			String:     "00000000",
			Size:       fontSize * ss.gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.RightHAlignment,
		},
		geometry.Point{
			X: bcPos.X - (size.Height * blockChainScale * ss.gs.Max * 0.75),
			Y: bcPos.Y,
		},
		bcColor,
		effects.Layer{Depth: -10},
	)

	// points display system
	world.AddSystem(ss.pointsDisplaySystem)

	// listen to points
	world.AddListener(ss.pointsListener)

	// text fade system
	world.AddSystem(ss.textFadeSystem)

	// listen to level events
	world.AddListener(ss.levelEvents)

	return err
}

func (ss *scoreSystem) pointsListener(world *goecs.World, signal interface{}, _ float32) error {
	if ss.end {
		return nil
	}
	switch e := signal.(type) {
	// we got points
	case PointsEvent:
		base := 0
		extra := 0

		if e.Total > 0 {
			// base points
			base = e.Total * pointPerBlock
			// multiply by 1 per each 4 blocks
			extra = e.Total / 4

			// if we have any extra add it
			if extra > 0 {
				ss.toAdd += base * extra
			} else {
				// add the base
				ss.toAdd += base
			}
		} else {
			base = e.Total * pointLosePerBlock
			ss.toSub += -base
		}

		ss.addFloatPoints(world, base, extra, e.At)
	}
	return nil
}

// update the points
func (ss *scoreSystem) pointsDisplaySystem(_ *goecs.World, delta float32) error {
	if ss.end {
		return nil
	}
	// if we have points to add
	if ss.toAdd > 0 {
		// add just some of them per tick, without overflow
		adding := int(math.Min(float64(pointsToAddPerSec*delta), float64(ss.toAdd)))
		ss.toAdd -= adding
		ss.total += adding
	}

	// if we have points to sub
	if ss.toSub > 0 {
		// add just some of them per tick, without overflow
		subbing := int(math.Min(float64(pointsToAddPerSec*delta), float64(ss.toSub)))
		ss.toSub -= subbing
		ss.total -= subbing
	}

	// if the score need to be update, update it
	if ss.lastScore != ss.total {
		ss.lastScore = ss.total
		text := ui.Get.Text(ss.textLabel)
		text.String = fmt.Sprintf("%d", ss.lastScore)

		ss.textLabel.Set(text)
	}

	return nil
}

// fate scroll text
func (ss *scoreSystem) textFadeSystem(world *goecs.World, delta float32) error {
	// get any text that is moving
	for it := world.Iterator(ui.TYPE.Text, color.TYPE.Solid, movement.Type, component.TYPE.FloatText); it != nil; it = it.Next() {
		ent := it.Value()
		// get the current color
		clr := color.Get.Solid(ent)
		// decrease
		a := int(clr.A) - int(textAlphaDecreesPerSec*delta)
		// if we need to dispear
		if a < 0 {
			// remove text
			if err := world.Remove(ent); err != nil {
				return err
			}
		} else {
			// update alpha
			clr.A = uint8(a)
			ent.Set(clr)
		}
	}

	return nil
}

func (ss *scoreSystem) addFloatPoints(world *goecs.World, base, extra int, at geometry.Point) {
	var text string
	txtColor := positiveColor
	if base > 0 {
		// format text for our floating text
		if extra > 1 {
			text = fmt.Sprintf("+%dx%d", base, extra)
		} else {
			text = fmt.Sprintf("+%d", base)
		}
	} else {
		text = fmt.Sprintf("%d", base)
		txtColor = negativeColor
	}

	// add the floating text
	world.AddEntity(
		ui.Text{
			String:     text,
			Size:       floatPointSize * ss.gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		at,
		txtColor,
		effects.Layer{Depth: -10},
		movement.Movement{
			Amount: geometry.Point{
				X: -textScrollSpeedX * ss.gs.Max,
				Y: -textScrollSpeedY * ss.gs.Max,
			},
		},
		component.FloatText{},
	)
}

func (ss *scoreSystem) levelEvents(world *goecs.World, signal interface{}, _ float32) error {
	switch signal.(type) {
	case winning.LevelEndEvent:
		ss.end = true

		ss.total += ss.toAdd
		ss.total -= ss.toSub
		ss.toAdd = 0
		ss.toSub = 0
		ss.lastScore = ss.total

		text := ui.Get.Text(ss.textLabel)
		text.String = fmt.Sprintf("%d", ss.lastScore)

		ss.textLabel.Set(text)
		return world.Signal(winning.FinalScoreEvent{Total: ss.total})
	}

	return nil
}

// System create the score system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	ss := scoreSystem{
		gs:        gs,
		dr:        dr,
		lastScore: -1,
	}
	return ss.load(engine)
}
