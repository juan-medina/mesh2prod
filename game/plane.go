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
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/animation"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/events"
)

// add the background
func addPlane(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

	// get the size of the first sprite for our plane
	if size, err = eng.GetSpriteSize(spriteSheet, fmt.Sprintf(gopherPlaneAnim, 1)); err != nil {
		return err
	}

	// calculate halve of the height
	halveHeight := (size.Height / 2) * planeScale

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
			amount: geometry.Point{},
			min: geometry.Point{
				X: 0,
				Y: halveHeight * gameScale.Point.X,
			},
			max: geometry.Point{
				X: designResolution.Width * gameScale.Point.X,
				Y: (designResolution.Height - halveHeight) * gameScale.Point.Y,
			},
		},
		effects.Layer{Depth: 0},
	)

	// add the keys listener
	world.AddListener(keyMoveListener)

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

			// if we have press the key calculate the speed
			if e.Status.Pressed {
				switch e.Key {
				case device.KeyUp:
					mov.amount.Y = -planeSpeed
				case device.KeyDown:
					mov.amount.Y = planeSpeed
				}
				// now we are animated faster
				anim.Speed = animSpeedFast
				// if not set speed to zero
			} else if e.Status.Released {
				mov.amount.Y = 0
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
