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

package background

import (
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/plane"
	"reflect"
)

const (
	bgLayer        = "resources/sprites/layer%d.png" // bg layers
	cloudLayers    = 6                               // number of cloud layers
	minCloudSpeed  = 200                             // min cloud speed
	cloudDiffSpeed = 40                              // difference of speed per layer
	parallaxEffect = 0.025                           // amount of parallax effect
)

var (
	cloudTransparency = color.White.Alpha(245) // our cloud transparency
)

type bgSystem struct {
	gs geometry.Scale
	dr geometry.Size
}

// add the background
func (bs bgSystem) load(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

	// add a gradient background
	world.AddEntity(
		shapes.Box{
			Size: geometry.Size{
				Width:  bs.dr.Width,
				Height: bs.dr.Height,
			},
			Scale: bs.gs.Min,
		},
		geometry.Point{},
		color.Gradient{
			From:      color.White,
			To:        color.SkyBlue,
			Direction: color.GradientVertical,
		},
	)

	flip := false
	// adding the clouds
	for ln := 0; ln < cloudLayers; ln++ {
		// get the file name
		lf := fmt.Sprintf(bgLayer, (ln>>1)+1)
		speed := -(minCloudSpeed + (cloudDiffSpeed * float32(cloudLayers-ln)))
		// load the sprite
		if err := eng.LoadSprite(lf, geometry.Point{X: 0, Y: 0}); err != nil {
			return err
		}
		if size, err = eng.GetSpriteSize("", lf); err != nil {
			return err
		}
		reset := size.Width * bs.gs.Point.X
		// add the first chunk
		world.AddEntity(
			sprite.Sprite{
				Name:  lf,
				Scale: bs.gs.Min,
				FlipX: flip,
			},
			geometry.Point{},
			movement.Movement{
				Amount: geometry.Point{
					X: speed,
					Y: 0,
				},
				Min: geometry.Point{
					X: -100000,
					Y: -100000,
				},
				Max: geometry.Point{
					X: 100000,
					Y: 100000,
				},
			},
			parallax{
				min:   -size.Width * bs.gs.Point.X,
				reset: reset,
				layer: ln,
			},
			cloudTransparency,
			effects.Layer{Depth: 1 + float32(ln)},
		)
		// add the second chunk
		world.AddEntity(
			sprite.Sprite{
				Name:  lf,
				Scale: bs.gs.Min,
				FlipX: !flip,
			},
			geometry.Point{X: reset},
			movement.Movement{
				Amount: geometry.Point{
					X: speed,
					Y: 0,
				},
				Min: geometry.Point{
					X: -100000,
					Y: -100000,
				},
				Max: geometry.Point{
					X: 100000,
					Y: 100000,
				},
			},
			parallax{
				min:   -size.Width * bs.gs.Point.X,
				reset: reset,
				layer: ln,
			},
			cloudTransparency,
			effects.Layer{Depth: 1 + float32(ln)},
		)
		flip = !flip
	}

	// add the reset system
	world.AddSystem(bs.resetSystem)

	// listen to plane changes
	world.AddListener(bs.planeChanges)

	return nil
}

// reset the layer if go off screen
func (bs *bgSystem) resetSystem(world *goecs.World, _ float32) error {
	// get our entities that has position and parallax
	for it := world.Iterator(geometry.TYPE.Point, parallaxType); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and movement
		pos := geometry.Get.Point(ent)
		par := ent.Get(parallaxType).(parallax)

		// if we are at our min reset
		if pos.X < par.min {
			pos.X = par.reset
			ent.Set(pos)
		}
	}
	return nil
}

//    0
//
//         300
//
//    600

// when plane changes position move the layers up or down a bit
func (bs *bgSystem) planeChanges(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		// calculate a shift from the plane position
		shift := (e.Pos.Y - (bs.dr.Height / 2 * bs.gs.Point.Y)) * parallaxEffect

		// get our entities that has position and parallax
		for it := world.Iterator(geometry.TYPE.Point, parallaxType); it != nil; it = it.Next() {
			// get the entity
			ent := it.Value()

			// get current position and movement
			pos := geometry.Get.Point(ent)
			par := ent.Get(parallaxType).(parallax)

			// update the position
			pos.Y = shift * float32(cloudLayers-par.layer+1)

			// update entity
			ent.Set(pos)
		}
	}
	return nil
}

type parallax struct {
	min   float32
	reset float32
	layer int
}

var parallaxType = reflect.TypeOf(parallax{})

// System creates the background system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	bs := bgSystem{
		gs: gs,
		dr: dr,
	}
	return bs.load(engine)
}
