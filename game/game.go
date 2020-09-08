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
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
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
	planeSpeed      = 900                                // plane speed
	animSpeedSlow   = 0.65                               // animation slow speed
	animSpeedFast   = 1                                  // animation fast speed
	meshSpriteAnim  = "box%d.png"                        // the mesh sprite
	meshScale       = 1                                  // mesh scale
	meshX           = 310                                // mesh scale
	meshSpeed       = float32(40)                        // mesh speed
	topMeshSpeed    = meshSpeed * 2                      // top mesh speed
	music           = "resources/music/loop.ogg"         // our game music
	bgLayer         = "resources/sprites/layer%d.png"    // bg layers
	cloudLayers     = 3                                  // number of cloud layers
	minCloudSpeed   = 200                                // min cloud speed
	cloudDiffSpeed  = 20                                 // difference of speed per layer
	parallaxEffect  = 0.010                              // amount of parallax effect
)

var (
	designResolution  = geometry.Size{Width: 1920, Height: 1080} // designResolution is how our game is designed
	gameScale         geometry.Scale                             // our game scale
	planeEnt          *goecs.Entity                              // our plane
	meshEnt           *goecs.Entity                              // our mesh
	cloudTransparency = color.White.Alpha(245)                   // our cloud transparency
)

// Load the game
func Load(eng *gosge.Engine) error {
	var err error
	var size geometry.Size

	// get the ECS world
	world := eng.World()

	// gameScale from the real screen size to our design resolution
	gameScale = eng.GetScreenSize().CalculateScale(designResolution)

	// load the music
	if err = eng.LoadMusic(music); err != nil {
		return err
	}

	// load the sprite sheet
	if err = eng.LoadSpriteSheet(spriteSheet); err != nil {
		return err
	}

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

	// add the follow system
	world.AddSystem(followSystem)

	// add the move system
	world.AddSystem(moveSystem)

	// add the parallaxSystem system
	world.AddSystem(parallaxSystem)

	// play the music
	return world.Signal(events.PlayMusicEvent{Name: music})
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
		pos.Y += mov.amount.Y * delta * gameScale.Point.X
		pos.X += mov.amount.X * delta * gameScale.Point.Y
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

// indicate how much we need to move
type movement struct {
	amount geometry.Point // how much we could move
	min    geometry.Point // min position that we could move
	max    geometry.Point // max position that we could move
}

var movementType = reflect.TypeOf(movement{})

type parallax struct {
	min   float32
	reset float32
	layer int
}

var parallaxType = reflect.TypeOf(parallax{})
