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

package target

import (
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
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/plane"
	"github.com/juan-medina/mesh2prod/game/winning"
	"math"
)

// logic constants
const (
	markSprite        = "mark.png"                 // mark sprite
	bulletSprite      = "bullet_%d.png"            // bullet sprite base
	bulletScale       = 0.25                       // scale for the bullet sprite
	bulletFrames      = 5                          // bullet frames
	bulletFramesDelay = 0.065                      // bullet frame delay
	bulletSpeed       = 600                        // bullet speed
	targetScale       = 0.5                        // block scale
	targetGapX        = 100                        // target gap from gun pos
	shotSound         = "resources/audio/shot.wav" // plane shot sound
)

var (
	bulletColor = color.Red.Alpha(180) // bullet color
)

type targetSystem struct {
	gs         geometry.Scale // game scale
	dr         geometry.Size  // design resolution
	targetSize geometry.Size  // block size
	gunPos     geometry.Point // plane gun position
	target     *goecs.Entity  // current target position
	line       *goecs.Entity  // target line
	end        bool
}

// load the system
func (gms *targetSystem) load(eng *gosge.Engine) error {
	var err error

	// get the block size
	if gms.targetSize, err = eng.GetSpriteSize(constants.SpriteSheet, markSprite); err != nil {
		return err
	}

	// laser sound
	if err = eng.LoadSound(shotSound); err != nil {
		return err
	}

	// get the world
	world := eng.World()

	// add the sprites from the current state
	gms.addSprites(world)

	// add the target system that target blocks
	world.AddSystem(gms.findTargetSystem)

	// listen to plane changes
	world.AddListener(gms.planeChanges, plane.PositionChangeEventType)

	// listen to keys
	world.AddListener(gms.keyListener, events.TYPE.KeyUpEvent)

	// listen to level events
	world.AddListener(gms.levelEvents, winning.LevelEndEventType)

	return nil
}

// add sprite from map state
func (gms *targetSystem) addSprites(world *goecs.World) {
	// add our target
	gms.target = world.AddEntity(
		geometry.Point{
			X: 0,
			Y: 0,
		},
		sprite.Sprite{
			Sheet: constants.SpriteSheet,
			Name:  markSprite,
			Scale: gms.gs.Max * targetScale,
		},
		effects.Layer{Depth: 0},
		effects.AlternateColor{
			From:  color.Red,
			To:    color.Red.Alpha(180),
			Time:  0.25,
			Delay: 0,
		},
	)

	// add the target line
	gms.line = world.AddEntity(
		geometry.Point{
			X: 0,
			Y: 0,
		},
		effects.AlternateColor{
			From:  color.Red.Alpha(60),
			To:    color.Red.Alpha(100),
			Time:  0.35,
			Delay: 0.35,
		},
		shapes.Line{
			To:        geometry.Point{},
			Thickness: 1.5 * gms.gs.Max,
		},
		effects.Layer{Depth: 0},
	)
}

// a system that target a block
func (gms *targetSystem) findTargetSystem(world *goecs.World, _ float32) error {
	if gms.end {
		return nil
	}
	// get the line from
	linePosFrom := geometry.Get.Point(gms.line)
	linePosFrom.X = gms.gunPos.X
	linePosFrom.Y = gms.gunPos.Y

	// get the line component
	line := shapes.Get.Line(gms.line)

	// try to find a target
	var found *goecs.Entity = nil

	// half size of block
	halfSize := gms.targetSize.Height * targetScale * gms.gs.Max * 0.5

	// screen width
	screenWith := gms.dr.Width * gms.gs.Point.X

	// close x
	closeX := float32(2000000)

	// search for the closets block
	for it := world.Iterator(component.TYPE.Block, sprite.TYPE, geometry.TYPE.Point); it != nil; it = it.Next() {
		// get entity values
		ent := it.Value()
		blockPos := geometry.Get.Point(ent)

		// if the block if off screen skipp it
		if blockPos.X > screenWith {
			continue
		}

		// difference in height
		diffY := float32(math.Abs(float64(blockPos.Y - gms.gunPos.Y)))

		// if we are under half block size
		if diffY <= halfSize {
			// diff in y
			diffX := blockPos.X - gms.gunPos.X
			// don't target to close things
			if diffX > targetGapX*gms.gs.Max {
				if diffX < closeX {
					closeX = diffX
					found = ent
				}
			}
		}
	}

	// if we have no a target
	if found == nil {
		// move target ouf ot screen
		gms.target.Set(geometry.Point{
			X: -1000,
			Y: -1000,
		})

		// move line straight from gun
		line.To = geometry.Point{
			X: gms.dr.Width * gms.gs.Max,
			Y: gms.gunPos.Y,
		}
	} else {
		pos := geometry.Get.Point(found)
		targetPos := geometry.Point{
			X: pos.X - (gms.targetSize.Width * gms.gs.Max * targetScale),
			Y: pos.Y,
		}
		gms.target.Set(targetPos)
		// calculate line pos
		line.To = geometry.Point{
			X: targetPos.X - (gms.targetSize.Width/2)*targetScale*gms.gs.Max,
			Y: targetPos.Y,
		}
	}

	// update line
	gms.line.Set(linePosFrom)
	gms.line.Set(line)

	return nil
}

// if the plane change position
func (gms *targetSystem) planeChanges(_ *goecs.World, signal interface{}, _ float32) error {
	if gms.end {
		return nil
	}
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		// store gun position
		gms.gunPos = e.Gun
	}
	return nil
}

// listen to keys
func (gms *targetSystem) keyListener(world *goecs.World, signal interface{}, _ float32) error {
	if gms.end {
		return nil
	}
	switch e := signal.(type) {
	// if we got a key up
	case events.KeyUpEvent:
		// if it space
		if e.Key == device.KeySpace {
			gms.createBullet(world)
		}
	}
	return nil
}

func (gms targetSystem) createBullet(world *goecs.World) {
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
						Scale:  gms.gs.Max * bulletScale,
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
					Y: velY * gms.gs.Max,
					X: bulletSpeed * gms.gs.Max,
				},
			},
			movement.Constrain{
				Min: geometry.Point{
					X: 0,
					Y: minY,
				},
				Max: geometry.Point{
					X: gms.dr.Width * gms.gs.Max,
					Y: maxY,
				},
			},
			bulletColor,
			component.Bullet{},
			effects.Layer{Depth: 0},
		)
		world.Signal(events.PlaySoundEvent{Name: shotSound, Volume: 1})
	}
}

func (gms *targetSystem) bulletSystem(world *goecs.World, _ float32) error {
	for it := world.Iterator(component.TYPE.Bullet, geometry.TYPE.Point); it != nil; it = it.Next() {
		bullet := it.Value()
		pos := geometry.Get.Point(bullet)
		if pos.X >= gms.dr.Width*gms.gs.Max {
			_ = world.Remove(bullet)
		}
	}
	return nil
}

func (gms *targetSystem) levelEvents(world *goecs.World, signal interface{}, _ float32) error {
	switch signal.(type) {
	case winning.LevelEndEvent:
		gms.end = true
		_ = world.Remove(gms.line)
		_ = world.Remove(gms.target)
	}

	return nil
}

// System create the target system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	gms := targetSystem{
		gs: gs,
		dr: dr,
	}
	return gms.load(engine)
}
