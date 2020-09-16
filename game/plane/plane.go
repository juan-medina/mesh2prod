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

package plane

import (
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/animation"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
)

const (
	gopherPlaneAnim = "gopher_plane_%d.png" // base animation for our gopher
	planeScale      = float32(0.5)          // plane scale
	planeX          = 720                   // plane X position
	planeSpeed      = 450                   // plane speed
	animSpeedSlow   = 0.65                  // animation slow speed
	animSpeedFast   = 1                     // animation fast speed
	joinShiftX      = 20                    // shift in X for the joint
	joinShiftY      = 5                     // shift in Y for the joint
)

type planeSystem struct {
	gs      geometry.Scale
	dr      geometry.Size
	plane   *goecs.Entity
	lastPos geometry.Point
	size    geometry.Size
}

// add the background
func (ps *planeSystem) load(eng *gosge.Engine) error {
	var err error

	// get the ECS world
	world := eng.World()

	// get the size of the first sprite for our plane
	if ps.size, err = eng.GetSpriteSize(constants.SpriteSheet, fmt.Sprintf(gopherPlaneAnim, 1)); err != nil {
		return err
	}

	// calculate halve of the height
	halveHeight := (ps.size.Height / 2) * planeScale

	// add our plane
	ps.plane = world.AddEntity(
		animation.Animation{
			Sequences: map[string]animation.Sequence{
				"flying": {
					Sheet:  constants.SpriteSheet,
					Base:   gopherPlaneAnim,
					Scale:  ps.gs.Min * planeScale,
					Frames: 2,
					Delay:  0.065,
				},
			},
			Current: "flying",
			Speed:   animSpeedSlow,
		},
		geometry.Point{
			X: planeX * ps.gs.Point.X,
			Y: ps.dr.Height / 2 * ps.gs.Point.Y,
		},
		movement.Movement{
			Amount: geometry.Point{},
		},
		movement.Constrain{
			Min: geometry.Point{
				X: 0,
				Y: halveHeight * ps.gs.Point.X,
			},
			Max: geometry.Point{
				X: ps.dr.Width * ps.gs.Point.X,
				Y: (ps.dr.Height - halveHeight) * ps.gs.Point.Y,
			},
		},
		effects.Layer{Depth: 0},
	)

	// add the keys listener
	world.AddListener(ps.keyMoveListener)

	// add system to notify the world of position changes
	world.AddSystem(ps.notifyPositionChanges)

	return nil
}

func (ps planeSystem) keyMoveListener(_ *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	// if we got a key event
	case events.KeyDownEvent:
		// if we have use the cursor up or down
		if e.Key == device.KeyUp || e.Key == device.KeyDown {
			// get the Movement and animation components
			mov := ps.plane.Get(movement.Type).(movement.Movement)
			anim := animation.Get.Animation(ps.plane)
			switch e.Key {
			case device.KeyUp:
				mov.Amount.Y = -planeSpeed
			case device.KeyDown:
				mov.Amount.Y = planeSpeed
			}
			// now we are animated faster
			anim.Speed = animSpeedFast

			// update the entity
			ps.plane.Set(mov)
			ps.plane.Set(anim)
		}
	case events.KeyUpEvent:
		// if we have use the cursor up or down
		if e.Key == device.KeyUp || e.Key == device.KeyDown {
			// get the Movement and animation components
			mov := ps.plane.Get(movement.Type).(movement.Movement)
			anim := animation.Get.Animation(ps.plane)

			// set speed to zero
			mov.Amount.Y = 0
			// now we are animated slower
			anim.Speed = animSpeedSlow

			// update the entity
			ps.plane.Set(mov)
			ps.plane.Set(anim)
		}
	}
	return nil
}

func (ps *planeSystem) notifyPositionChanges(world *goecs.World, _ float32) error {
	current := geometry.Get.Point(ps.plane)

	if current.X != ps.lastPos.X || current.Y != ps.lastPos.Y {
		ps.lastPos = current
		joint := geometry.Point{
			X: current.X - (((ps.size.Width / 2) - joinShiftX) * planeScale * ps.gs.Point.X),
			Y: current.Y - (joinShiftY * planeScale * ps.gs.Point.Y),
		}
		return world.Signal(PositionChangeEvent{Pos: current, Joint: joint})
	}

	return nil
}

// System create a plane system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	ps := planeSystem{
		gs:    gs,
		dr:    dr,
		plane: nil,
	}

	return ps.load(engine)
}

// PositionChangeEvent notify others that the plane has change position
type PositionChangeEvent struct {
	Pos   geometry.Point // Pos is where ir our plane is
	Joint geometry.Point // Joint is the joint point for our plane
}
