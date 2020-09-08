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
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"reflect"
)

// add the background
func addBackground(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

	// add a gradient background
	world.AddEntity(
		shapes.Box{
			Size: geometry.Size{
				Width:  designResolution.Width,
				Height: designResolution.Height,
			},
			Scale: gameScale.Min,
		},
		geometry.Point{},
		color.Gradient{
			From:      color.White,
			To:        color.SkyBlue,
			Direction: color.GradientVertical,
		},
	)

	// adding the clouds
	for ln := 1; ln <= cloudLayers; ln++ {
		// get the file name
		lf := fmt.Sprintf(bgLayer, ln)
		speed := -(minCloudSpeed + (cloudDiffSpeed * float32(cloudLayers-ln)))
		// load the sprite
		if err := eng.LoadSprite(lf, geometry.Point{X: 0, Y: 0}); err != nil {
			return err
		}
		if size, err = eng.GetSpriteSize("", lf); err != nil {
			return err
		}
		reset := size.Width * gameScale.Point.X
		// add the first chunk
		world.AddEntity(
			sprite.Sprite{
				Name:  lf,
				Scale: gameScale.Min,
			},
			geometry.Point{},
			movement{
				amount: geometry.Point{
					X: speed,
					Y: 0,
				},
				min: geometry.Point{
					X: -100000,
					Y: 0,
				},
				max: geometry.Point{
					X: 100000,
					Y: 100000,
				},
			},
			parallax{
				min:   -size.Width * gameScale.Point.X,
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
				Scale: gameScale.Min,
				FlipX: true,
			},
			geometry.Point{X: reset},
			movement{
				amount: geometry.Point{
					X: speed,
					Y: 0,
				},
				min: geometry.Point{
					X: -100000,
					Y: 0,
				},
				max: geometry.Point{
					X: 100000,
					Y: 100000,
				},
			},
			parallax{
				min:   -size.Width * gameScale.Point.X,
				reset: reset,
				layer: ln,
			},
			cloudTransparency,
			effects.Layer{Depth: 1 + float32(ln)},
		)
	}

	// add the parallaxSystem system
	world.AddSystem(parallaxSystem)

	return nil
}

func parallaxSystem(world *goecs.World, _ float32) error {
	// get our entities that has position and parallax
	for it := world.Iterator(geometry.TYPE.Point, parallaxType); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and movement
		pos := geometry.Get.Point(ent)
		par := ent.Get(parallaxType).(parallax)

		// if we are at our mine reset
		if pos.X < par.min {
			pos.X = par.reset
		}

		planePos := geometry.Get.Point(planeEnt)

		shift := ((designResolution.Height / 2 * gameScale.Point.Y) - planePos.Y) * parallaxEffect * gameScale.Min

		pos.Y = shift * float32(cloudLayers-par.layer+1)

		ent.Set(pos)
	}
	return nil
}

type parallax struct {
	min   float32
	reset float32
	layer int
}

var parallaxType = reflect.TypeOf(parallax{})
