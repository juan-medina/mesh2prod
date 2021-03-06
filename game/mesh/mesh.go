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
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/mesh2prod/game/component"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/plane"
	"github.com/juan-medina/mesh2prod/game/winning"
)

const (
	animSpeedSlow    = 0.65         // animation slow speed
	meshSpriteAnim   = "box%d.png"  // the mesh sprite
	meshScale        = 0.5          // mesh scale
	meshX            = 10           // mesh scale
	meshSpeed        = float32(200) // mesh speed
	topMeshSpeed     = float32(250) // top mesh speed
	joinShiftX       = 5            // shift X for the joint
	joinShiftYTop    = 130          // shift Y for the top joint
	joinShiftYBottom = 170          // shift Y for the bottom joint
	lineThickness    = 5            // the line thickness
	meshScrollSpeedX = 25           // mesh scroll x (match block scroll)
)

type meshSystem struct {
	gs       geometry.Scale
	dr       geometry.Size
	mesh     *goecs.Entity
	planePos geometry.Point
	line     [2]*goecs.Entity
	size     geometry.Size
	end      bool
}

// add the background
func (ms *meshSystem) load(eng *gosge.Engine) error {
	var err error

	// get the ECS world
	world := eng.World()

	// get the size of the mesh
	if ms.size, err = eng.GetSpriteSize(constants.SpriteSheet, fmt.Sprintf(meshSpriteAnim, 1)); err != nil {
		return err
	}

	// calculate halve of the height
	halveHeight := (ms.size.Height / 2) * meshScale

	// add the mesh
	ms.mesh = world.AddEntity(
		animation.Animation{
			Sequences: map[string]animation.Sequence{
				"flying": {
					Sheet:  constants.SpriteSheet,
					Base:   meshSpriteAnim,
					Scale:  ms.gs.Max * meshScale,
					Frames: 2,
					Delay:  0.065,
				},
			},
			Current: "flying",
			Speed:   animSpeedSlow,
		},
		geometry.Point{
			X: (ms.size.Width / 2 * meshScale * ms.gs.Max) + meshX*ms.gs.Max,
			Y: ms.dr.Height / 2 * ms.gs.Max,
		},
		movement.Movement{
			Amount: geometry.Point{
				X: 0,
				Y: 100,
			},
		},
		movement.Constrain{
			Min: geometry.Point{
				X: 0,
				Y: halveHeight * ms.gs.Max,
			},
			Max: geometry.Point{
				X: ms.dr.Width * ms.gs.Max,
				Y: (ms.dr.Height - halveHeight) * ms.gs.Max,
			},
		},
		component.Mesh{},
		effects.Layer{Depth: 0},
	)

	// create the two lines
	for ln := 0; ln < 2; ln++ {
		ms.line[ln] = world.AddEntity(
			shapes.Line{
				To:        geometry.Point{},
				Thickness: lineThickness * ms.gs.Max,
			},
			geometry.Point{},
			effects.Layer{Depth: 0.5},
			color.Gray,
		)
	}

	// add the follow system
	world.AddSystem(ms.followSystem)

	// listen to plane changes
	world.AddListener(ms.planeChanges, plane.PositionChangeEventType)

	// listen to level events
	world.AddListener(ms.levelEvents, winning.LevelEndEventType)

	return nil
}

// follow system
func (ms *meshSystem) followSystem(_ *goecs.World, delta float32) error {
	if ms.end {
		return nil
	}
	// get mesh component
	meshPos := geometry.Get.Point(ms.mesh)
	mov := ms.mesh.Get(movement.Type).(movement.Movement)

	// calculate difference
	diffY := ms.planePos.Y - meshPos.Y

	// increase Movement up or down
	mov.Amount.Y = diffY * meshSpeed * ms.gs.Max * delta

	// clamp speed
	if mov.Amount.Y > topMeshSpeed {
		mov.Amount.Y = topMeshSpeed
	} else if mov.Amount.Y < -topMeshSpeed {
		mov.Amount.Y = -topMeshSpeed
	}

	// update the mesh Movement
	ms.mesh.Set(mov)

	// we will calculate the line from position
	var linePos geometry.Point

	// calculate X, the same for both lines
	linePos.X = meshPos.X + (((ms.size.Width / 2) - joinShiftX) * meshScale * ms.gs.Max)

	// top line
	linePos.Y = meshPos.Y - (joinShiftYTop * meshScale * ms.gs.Max)
	ms.line[0].Set(linePos)

	// bottom line
	linePos.Y = meshPos.Y + (joinShiftYBottom * meshScale * ms.gs.Max)
	ms.line[1].Set(linePos)

	return nil
}

// when plane changes save it position
func (ms *meshSystem) planeChanges(_ *goecs.World, signal interface{}, _ float32) error {
	if ms.end {
		return nil
	}
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		ms.planePos = e.Pos
		// go the lines and update the to, same for both
		for ln := 0; ln < 2; ln++ {
			line := shapes.Get.Line(ms.line[ln])
			line.To = e.Joint // we will use the joint position send by the plane
			ms.line[ln].Set(line)
		}
	}
	return nil
}

func (ms *meshSystem) levelEvents(world *goecs.World, signal interface{}, _ float32) error {
	switch signal.(type) {
	case winning.LevelEndEvent:
		ms.end = true
		for ln := 0; ln < 2; ln++ {
			_ = world.Remove(ms.line[ln])
			ms.mesh.Set(movement.Movement{Amount: geometry.Point{
				X: -meshScrollSpeedX * ms.gs.Max,
				Y: 0,
			}})
			ms.mesh.Remove(movement.ConstrainType)
		}
	}

	return nil
}

// System creates the mesh system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	bs := meshSystem{
		gs: gs,
		dr: dr,
	}
	return bs.load(engine)
}
