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
	"github.com/juan-medina/mesh2prod/game/plane"
	"math/rand"
	"reflect"
)

const (
	cloudLayers    = 6     // number of cloud layers
	numClouds      = 20    // number of clouds
	minCloudSpeed  = 200   // Min cloud speed
	cloudDiffSpeed = 40    // difference of speed per Layer
	parallaxEffect = 0.045 // amount of Parallax effect
)

var (
	cloudTransparency = color.White.Alpha(235) // our cloud transparency

	scalePerLayer = [...]float32{0.8, 0.7, 0.6, 0.5, 0.4, 0.3}
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
		// for each number off cloud
		for cn := 0; cn < numClouds; cn++ {
			// add a cloud
			ent = world.AddEntity(
				cloudTransparency,
				effects.Layer{Depth: 1 + float32(ln)},
			)
			// set it random
			if err = bs.updateCloud(ent, 0, ln); err != nil {
				return err
			}

			// add another cloud
			ent = world.AddEntity(
				cloudTransparency,
				effects.Layer{Depth: 1 + float32(ln)},
			)
			// set it random off screen
			if err = bs.updateCloud(ent, bs.dr.Width, ln); err != nil {
				return err
			}
		}
	}

	// add the reset system
	world.AddSystem(bs.resetSystem)

	// listen to plane changes
	world.AddListener(bs.planeChanges)

	return nil
}

func (bs *bgSystem) updateCloud(ent *goecs.Entity, from float32, ln int) error {
	// get a random sprite
	spn := rand.Intn(3) + 1
	sf := fmt.Sprintf("cloud%d.PNG", spn)

	// set the scale according to the layer
	scale := scalePerLayer[ln]

	// sprite component
	spr := sprite.Sprite{
		Sheet: constants.SpriteSheet,
		Name:  sf,
		Scale: bs.gs.Min * scale,
		FlipX: rand.Intn(2) == 0,
	}

	// speed base on the layer
	speed := -(minCloudSpeed + (cloudDiffSpeed * (cloudLayers - float32(ln))))

	// movement component
	mov := movement.Movement{
		Amount: geometry.Point{
			X: speed,
			Y: 0,
		},
	}

	// calculate position
	y := (rand.Float32() * bs.dr.Height) * bs.gs.Point.Y
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
	par := Parallax{
		Min:   -((size.Width / 2) * bs.gs.Point.X * scale),
		Layer: ln,
		Y:     y,
	}

	// update entity
	ent.Set(spr)
	ent.Set(mov)
	ent.Set(pos)
	ent.Set(par)

	return nil
}

// reset the Layer if go off screen
func (bs *bgSystem) resetSystem(world *goecs.World, _ float32) error {
	// get our entities that has position and Parallax
	for it := world.Iterator(geometry.TYPE.Point, ParallaxType); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and movement
		pos := geometry.Get.Point(ent)
		par := ent.Get(ParallaxType).(Parallax)

		// if we are at our Min reset
		if pos.X < par.Min {
			if err := bs.updateCloud(ent, bs.dr.Width*bs.gs.Min, par.Layer); err != nil {
				return err
			}
		}
	}
	return nil
}

// when plane changes position move the layers up or down a bit
func (bs *bgSystem) planeChanges(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case plane.PositionChangeEvent:
		// calculate a shift from the plane position
		shift := (e.Pos.Y - (bs.dr.Height / 2 * bs.gs.Point.Y)) * parallaxEffect

		// get our entities that has position and Parallax
		for it := world.Iterator(geometry.TYPE.Point, ParallaxType); it != nil; it = it.Next() {
			// get the entity
			ent := it.Value()

			// get current position and movement
			pos := geometry.Get.Point(ent)
			par := ent.Get(ParallaxType).(Parallax)

			// update the position
			pos.Y = par.Y - (shift * float32(cloudLayers-par.Layer+1))

			// update entity
			ent.Set(pos)
		}
	}
	return nil
}

// Parallax represent the Parallax effect for our object
type Parallax struct {
	Min   float32 // Min position of this element to be off screen
	Layer int     // Layer for the Parallax effect
	Y     float32 // the original Y for the element
}

// ParallaxType is the reflect.Type of Parallax
var ParallaxType = reflect.TypeOf(Parallax{})

// System creates the background system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	bs := bgSystem{
		gs: gs,
		dr: dr,
	}
	return bs.load(engine)
}
