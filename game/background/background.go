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
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/movement"
	"math/rand"
	"reflect"
)

const (
	cloudLayers    = 8   // number of cloud layers
	numClouds      = 10  // number of clouds
	minCloudSpeed  = 200 // Min cloud speed
	cloudDiffSpeed = 60  // difference of speed per Layer
)

type bgSystem struct {
	gs  geometry.Scale
	dr  geometry.Size
	eng *gosge.Engine
}

// add the background
func (bs *bgSystem) load(eng *gosge.Engine) error {
	var err error

	bs.eng = eng

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
	var ent *goecs.Entity

	// for each layer
	for ln := 0; ln < cloudLayers; ln++ {
		alpha := 180 - uint8((float32(ln)/cloudLayers)*100)
		cloudAlpha := color.White.Alpha(alpha)
		// for each number off cloud
		for cn := 0; cn < numClouds; cn++ {
			// add a cloud
			ent = world.AddEntity(
				cloudAlpha,
				effects.Layer{Depth: 1 + float32(ln)},
			)
			// set it random
			if err = bs.resetCloud(ent, 0); err != nil {
				return err
			}

			// add another cloud off-screen
			ent = world.AddEntity(
				cloudAlpha,
				effects.Layer{Depth: 1 + float32(ln)},
			)
			// set it random off screen
			if err = bs.resetCloud(ent, bs.dr.Width); err != nil {
				return err
			}
		}
	}

	// add the reset system
	world.AddSystem(bs.resetSystem)

	return nil
}

func (bs *bgSystem) resetCloud(ent *goecs.Entity, from float32) error {
	// get a random sprite
	spn := rand.Intn(3) + 1
	sf := fmt.Sprintf("cloud%d.PNG", spn)

	// get the layer number
	layer := effects.Get.Layer(ent)
	ln := layer.Depth - 1

	// set the scale according to the layer
	scale := 1.5 - (ln/cloudLayers)*1.5

	// sprite component
	spr := sprite.Sprite{
		Sheet: constants.SpriteSheet,
		Name:  sf,
		Scale: bs.gs.Min * scale,
		FlipX: rand.Intn(2) == 0, // random flipped horizontally
	}

	// speed base on the layer
	speed := -(minCloudSpeed + (cloudDiffSpeed * (cloudLayers - ln)))

	// movement component
	mov := movement.Movement{
		Amount: geometry.Point{
			X: speed,
			Y: 0,
		},
	}
	y := float32(0)

	top := rand.Intn(2) == 0

	if top {
		y = ((bs.dr.Height / 2) * bs.gs.Point.Y) * scale
		y += bs.dr.Height / 6
	} else {
		y = bs.dr.Height * bs.gs.Point.Y
		y -= bs.dr.Height / 6
		y -= ((bs.dr.Height / 2) * bs.gs.Point.Y) * scale
	}

	// calculate position
	pos := geometry.Point{
		X: (rand.Float32()*bs.dr.Width + from) * bs.gs.Point.X,
		Y: y,
	}

	var size geometry.Size
	var err error

	//get sprite size
	if size, err = bs.eng.GetSpriteSize(constants.SpriteSheet, sf); err != nil {
		return err
	}

	// parallax component
	rst := Reset{At: -((size.Width / 2) * bs.gs.Point.X * scale)}

	// update entity
	ent.Set(spr)
	ent.Set(mov)
	ent.Set(pos)
	ent.Set(rst)

	return nil
}

// reset the Layer if go off screen
func (bs *bgSystem) resetSystem(world *goecs.World, _ float32) error {
	// get our entities that has position and Reset
	for it := world.Iterator(geometry.TYPE.Point, ResetType); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and reset
		pos := geometry.Get.Point(ent)
		rst := ent.Get(ResetType).(Reset)
		spr := sprite.Get(ent)

		// if we are at our Min reset
		if pos.X < rst.At {
			if size, err := bs.eng.GetSpriteSize(spr.Sheet, spr.Name); err == nil {
				ss := (size.Width / 2) * spr.Scale
				if err := bs.resetCloud(ent, (bs.dr.Width+ss)*bs.gs.Min); err != nil {
					return err
				}
			} else {
				return err
			}

		}
	}
	return nil
}

// Reset represent the Reset effect for our object
type Reset struct {
	At float32 // Min position of this element to be off screen
}

// ResetType is the reflect.Type of Reset
var ResetType = reflect.TypeOf(Reset{})

// System creates the background system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	bs := bgSystem{
		gs: gs,
		dr: dr,
	}
	return bs.load(engine)
}
