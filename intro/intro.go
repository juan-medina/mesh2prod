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

package intro

import (
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/gosge/events"
	"reflect"
)

// intro logic constants
const (
	bgSprite = "resources/intro/newolds.png"
	font     = "resources/fonts/go_regular.fnt" // our version text font
	fontSize = 30                               // text font size
)

// Stage the intro
func Stage(eng *gosge.Engine) error {
	var err error
	eng.DisableExitKey()

	// preload sprite
	if err = eng.LoadSprite(bgSprite, geometry.Point{X: 0.5, Y: 0.5}); err != nil {
		return err
	}

	// preload font
	if err = eng.LoadFont(font); err != nil {
		return err
	}

	// design resolution is how our game is designed
	dr := geometry.Size{Width: 1920, Height: 1080}

	// game scale from the real screen size to our design resolution
	gs := eng.GetScreenSize().CalculateScale(dr)

	// get the ECS world
	world := eng.World()

	// add the background
	addBackground(world, dr, gs)

	// add the version
	addVersion(world, dr, gs, eng.GetSettings().GetString("version", ""))

	// add the fade system
	world.AddSystem(fadeSystem)

	// add the fade off listener
	world.AddListener(fadeOffListener, fadeOffEventType)

	return err
}

func fadeOffListener(world *goecs.World, signal interface{}, _ float32) error {
	switch signal.(type) {
	case fadeOffEvent:
		for it := world.Iterator(color.TYPE.Solid); it != nil; it = it.Next() {
			ent := it.Value()
			ent.Add(
				fadeTo{
					alpha:  0,
					time:   2,
					signal: events.ChangeGameStage{Stage: "menu"},
				},
			)
		}
	}
	return nil
}

func fadeSystem(world *goecs.World, delta float32) error {
	for it := world.Iterator(fadeToType, color.TYPE.Solid, sprite.TYPE); it != nil; it = it.Next() {
		ent := it.Value()
		clr := color.Get.Solid(ent)
		fad := ent.Get(fadeToType).(fadeTo)

		if fad.current == 0 {
			fad.oriAlpha = clr.A
		}
		fad.current += delta
		if fad.current > fad.time {
			ent.Remove(fadeToType)
			ent.Set(clr.Alpha(fad.alpha))
			world.Signal(fad.signal)
			continue
		}
		clr = clr.Alpha(fad.oriAlpha).Blend(clr.Alpha(fad.alpha), fad.current/fad.time)

		ent.Set(clr)
		ent.Set(fad)
	}
	return nil
}

func addVersion(world *goecs.World, dr geometry.Size, gs geometry.Scale, version string) {
	versionPos := geometry.Point{
		X: (dr.Width - 10) * gs.Point.X,
		Y: (dr.Height - 10) * gs.Point.Y,
	}
	world.AddEntity(
		ui.Text{
			String:     version,
			Size:       fontSize * gs.Max,
			Font:       font,
			VAlignment: ui.BottomVAlignment,
			HAlignment: ui.RightHAlignment,
		},
		versionPos,
		color.White,
	)
}

func addBackground(world *goecs.World, dr geometry.Size, gs geometry.Scale) {
	// add logo
	world.AddEntity(
		sprite.Sprite{
			Name:  bgSprite,
			Scale: gs.Point.X,
		},
		geometry.Point{
			X: dr.Width * gs.Point.X * 0.5,
			Y: dr.Height * gs.Point.Y * 0.5,
		},
		color.White.Alpha(0),
		fadeTo{
			alpha: 255,
			time:  3,
			signal: events.DelaySignal{
				Signal: fadeOffEvent{},
				Time:   2,
			},
		},
	)
}

type fadeTo struct {
	alpha    byte
	time     float32
	signal   interface{}
	current  float32
	oriAlpha byte
}

var fadeToType = reflect.TypeOf(fadeTo{})

type fadeOffEvent struct{}

var fadeOffEventType = reflect.TypeOf(fadeOffEvent{})
