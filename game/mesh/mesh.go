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

package mesh

import (
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/animation"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/mesh2prod/game/movement"
)

const (
	spriteSheet    = "resources/sprites/mesh2prod.json" // game sprite sheet
	animSpeedSlow  = 0.65                               // animation slow speed
	meshSpriteAnim = "box%d.png"                        // the mesh sprite
	meshScale      = 1                                  // mesh scale
	meshX          = 310                                // mesh scale
	meshSpeed      = float32(40)                        // mesh speed
	topMeshSpeed   = meshSpeed * 2                      // top mesh speed
)

type meshSystem struct {
	gs    geometry.Scale
	dr    geometry.Size
	plane *goecs.Entity
	mesh  *goecs.Entity
}

// add the background
func (ms *meshSystem) load(eng *gosge.Engine) error {
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
	ms.mesh = world.AddEntity(
		animation.Animation{
			Sequences: map[string]animation.Sequence{
				"flying": {
					Sheet:  spriteSheet,
					Base:   meshSpriteAnim,
					Scale:  ms.gs.Min * meshScale,
					Frames: 2,
					Delay:  0.065,
				},
			},
			Current: "flying",
			Speed:   animSpeedSlow,
		},
		geometry.Point{
			X: meshX * ms.gs.Point.X,
			Y: ms.dr.Height / 2 * ms.gs.Point.Y,
		},
		movement.Movement{
			Amount: geometry.Point{
				X: 0,
				Y: 100,
			},
			Min: geometry.Point{
				X: 0,
				Y: halveHeight * ms.gs.Point.X,
			},
			Max: geometry.Point{
				X: ms.dr.Width * ms.gs.Point.X,
				Y: (ms.dr.Height - halveHeight) * ms.gs.Point.Y,
			},
		},
		effects.Layer{Depth: 0},
	)

	// add the follow system
	world.AddSystem(ms.followSystem)

	return nil
}

// follow system
func (ms meshSystem) followSystem(_ *goecs.World, delta float32) error {
	// get plane components
	planePos := geometry.Get.Point(ms.plane)

	// get mesh component
	meshPos := geometry.Get.Point(ms.mesh)
	mov := ms.mesh.Get(movement.Type).(movement.Movement)

	// calculate difference
	diffY := planePos.Y - meshPos.Y

	// increase Movement up or down
	if diffY > 0 {
		mov.Amount.Y += meshSpeed * delta
	} else {
		mov.Amount.Y += -meshSpeed * delta
	}

	// clamp speed
	if mov.Amount.Y > topMeshSpeed {
		mov.Amount.Y = topMeshSpeed
	} else if mov.Amount.Y < -topMeshSpeed {
		mov.Amount.Y = -topMeshSpeed
	}

	// update the mesh Movement
	ms.mesh.Set(mov)

	return nil
}

// System creates the mesh system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size, plane *goecs.Entity) error {
	bs := meshSystem{
		gs:    gs,
		dr:    dr,
		plane: plane,
	}
	return bs.load(engine)
}
