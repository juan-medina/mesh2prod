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

package game

import (
	"fmt"
	"reflect"

	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/animation"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/events"
)

const (
	spriteSheet     = "resources/sprites/mesh2prod.json" // game sprite sheet
	gopherPlaneAnim = "gopher_plane_%d.png"              // base animation for our gopher
	planeScale      = float32(0.5)                       // plane scale
	planeX          = 720                                // plane X position
	planeSpeed      = 1000                               // plane speed
	animSpeedSlow   = 0.65                               // animation slow speed
	animSpeedFast   = 1                                  // animation fast speed
	meshSpriteAnim  = "box%d.png"                        // the mesh sprite
	meshScale       = 1                                  // mesh scale
	meshX           = 310                                // mesh scale
	meshSpeed       = float32(100)                       // mesh speed
	topMeshSpeed    = meshSpeed * 2                      // top mesh speed
)

var (
	// designResolution is how our game is designed
	designResolution = geometry.Size{Width: 1920, Height: 1080}

	planeEnt *goecs.Entity // our plane
	meshEnt  *goecs.Entity // our mesh
)

// Load the game
func Load(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

	// gameScale from the real screen size to our design resolution
	gameScale := eng.GetScreenSize().CalculateScale(designResolution)

	// load the sprite sheet
	if err = eng.LoadSpriteSheet(spriteSheet); err != nil {
		return err
	}

	// get the size of the mesh
	if size, err = eng.GetSpriteSize(spriteSheet, fmt.Sprintf(meshSpriteAnim, 1)); err != nil {
		return err
	}

	// calculate halve of the height
	halveHeight := (size.Height / 2) * meshScale

	// add the mesh
	meshEnt = world.AddEntity(
		animation.Animation{
			Sequences: map[string]animation.Sequence{
				"flying": {
					Sheet:  spriteSheet,
					Base:   meshSpriteAnim,
					Scale:  gameScale.Min * meshScale,
					Frames: 2,
					Delay:  0.065,
				},
			},
			Current: "flying",
			Speed:   animSpeedSlow,
		},
		geometry.Point{
			X: meshX * gameScale.Point.X,
			Y: designResolution.Height / 2 * gameScale.Point.Y,
		},
		movement{
			amount: 100,
			min: geometry.Point{
				X: 0,
				Y: halveHeight * gameScale.Point.X,
			},
			max: geometry.Point{
				X: designResolution.Width * gameScale.Point.X,
				Y: (designResolution.Height - halveHeight) * gameScale.Point.Y,
			},
		},
	)

	// get the size of the first sprite for our plane
	if size, err = eng.GetSpriteSize(spriteSheet, fmt.Sprintf(gopherPlaneAnim, 1)); err != nil {
		return err
	}

	// calculate halve of the height
	halveHeight = (size.Height / 2) * planeScale

	// add our plane
	planeEnt = world.AddEntity(
		animation.Animation{
			Sequences: map[string]animation.Sequence{
				"flying": {
					Sheet:  spriteSheet,
					Base:   gopherPlaneAnim,
					Scale:  gameScale.Min * planeScale,
					Frames: 2,
					Delay:  0.065,
				},
			},
			Current: "flying",
			Speed:   animSpeedSlow,
		},
		geometry.Point{
			X: planeX * gameScale.Point.X,
			Y: designResolution.Height / 2 * gameScale.Point.Y,
		},
		movement{
			amount: 0,
			min: geometry.Point{
				X: 0,
				Y: halveHeight * gameScale.Point.X,
			},
			max: geometry.Point{
				X: designResolution.Width * gameScale.Point.X,
				Y: (designResolution.Height - halveHeight) * gameScale.Point.Y,
			},
		},
	)

	// add the keys listener
	world.AddListener(keyMoveListener)

	// add the follow system
	world.AddSystem(followSystem)

	// add the move system
	world.AddSystem(moveSystem)

	return nil
}

// move system
func moveSystem(world *goecs.World, delta float32) error {
	// move anything that has a position and movement
	for it := world.Iterator(geometry.TYPE.Point, movementType); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and movement
		pos := geometry.Get.Point(ent)
		mov := ent.Get(movementType).(movement)

		// increment position and clamp to the min/max
		pos.Y += mov.amount * delta
		pos.Clamp(mov.min, mov.max)

		// update entity
		ent.Set(pos)
	}

	return nil
}

func keyMoveListener(_ *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	// if we got a key event
	case events.KeyEvent:
		// if we have use the cursor up or down
		if e.Key == device.KeyUp || e.Key == device.KeyDown {
			// get the movement and animation components
			mov := planeEnt.Get(movementType).(movement)
			anim := animation.Get.Animation(planeEnt)

			// if we have pres the key calculate the speed
			if e.Status.Pressed {
				switch e.Key {
				case device.KeyUp:
					mov.amount = -planeSpeed
				case device.KeyDown:
					mov.amount = planeSpeed
				}
				// now we are animated faster
				anim.Speed = animSpeedFast
				// if not set speed to zero
			} else if e.Status.Released {
				mov.amount = 0
				// now we are animated slower
				anim.Speed = animSpeedSlow
			}
			// update the entity
			planeEnt.Set(mov)
			planeEnt.Set(anim)
		}
	}
	return nil
}

// follow system
func followSystem(_ *goecs.World, delta float32) error {
	// get components
	planePos := geometry.Get.Point(planeEnt)
	meshPos := geometry.Get.Point(meshEnt)
	mov := meshEnt.Get(movementType).(movement)

	// calculate difference
	diffY := planePos.Y - meshPos.Y

	// increase movement up or down
	if diffY > 0 {
		mov.amount += meshSpeed * delta
	} else {
		mov.amount += -meshSpeed * delta
	}

	// clamp speed
	if mov.amount > topMeshSpeed {
		mov.amount = topMeshSpeed
	} else if mov.amount < -topMeshSpeed {
		mov.amount = -topMeshSpeed
	}

	// update the mesh movement
	meshEnt.Set(mov)

	return nil
}

// indicate how much we need to move
type movement struct {
	amount float32        // how much we could move
	min    geometry.Point // min position that we could move
	max    geometry.Point // max position that we could move
}

var movementType = reflect.TypeOf(movement{})
