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
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
)

// add the background
func addMesh(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

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
			amount: geometry.Point{
				X: 0,
				Y: 100,
			},
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

	// add the follow system
	world.AddSystem(followSystem)

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
		mov.amount.Y += meshSpeed * delta
	} else {
		mov.amount.Y += -meshSpeed * delta
	}

	// clamp speed
	if mov.amount.Y > topMeshSpeed {
		mov.amount.Y = topMeshSpeed
	} else if mov.amount.Y < -topMeshSpeed {
		mov.amount.Y = -topMeshSpeed
	}

	// update the mesh movement
	meshEnt.Set(mov)

	return nil
}
