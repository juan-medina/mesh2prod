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

package menu

import (
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/gosge/events"
)

const (
	uiSheet           = "resources/sprites/ui.json"
	gamerSprite       = "gamer.png"
	logoSprite        = "logo.png"
	clickSound        = "resources/audio/click.wav"      // button click sound
	shadowExtraWidth  = 3                                // the x offset for the buttons shadow
	shadowExtraHeight = 3                                // the y offset for the buttons shadow
	font              = "resources/fonts/go_regular.fnt" // our message text font
	fontBigSize       = 60                               // big text font size
	fontSmallSize     = 30                               // small text font size
	logoScale         = 0.75                             // logo scale
	buttonExtraWidth  = 0.15                             // the additional width for a button si it is not only the text size
	buttonExtraHeight = 0.20                             // the additional width for a button si it is not only the text size
	music             = "resources/music/menu/Of Far Different Nature - Adventure Begins (CC-BY).ogg"
)

// Stage the menu
func Stage(eng *gosge.Engine) error {
	var err error

	eng.SetExitKey(device.KeyEscape)

	// preload music
	if err = eng.LoadMusic(music); err != nil {
		return err
	}
	// load the sprite sheet
	if err = eng.LoadSpriteSheet(uiSheet); err != nil {
		return err
	}

	// pre-load click sound
	if err = eng.LoadSound(clickSound); err != nil {
		return err
	}

	// pre-load font
	if err = eng.LoadFont(font); err != nil {
		return err
	}

	// design resolution is how our game is designed
	dr := geometry.Size{Width: 1920, Height: 1080}

	// game scale from the real screen size to our design resolution
	gs := eng.GetScreenSize().CalculateScale(dr)

	// get the ECS world
	world := eng.World()

	// add a gradient background
	world.AddEntity(
		shapes.SolidBox{
			Size: geometry.Size{
				Width:  dr.Width,
				Height: dr.Height / 2,
			},
			Scale: gs.Max,
		},
		geometry.Point{},
		color.Gradient{
			From:      color.White,
			To:        color.SkyBlue,
			Direction: color.GradientVertical,
		},
	)
	// add a gradient background
	world.AddEntity(
		shapes.SolidBox{
			Size: geometry.Size{
				Width:  dr.Width,
				Height: dr.Height * 0.5,
			},
			Scale: gs.Max,
		},
		geometry.Point{
			Y: dr.Height * 0.5 * gs.Max,
		},
		color.Gradient{
			From:      color.SkyBlue,
			To:        color.Blue,
			Direction: color.GradientVertical,
		},
	)

	// gopher sprite
	world.AddEntity(
		sprite.Sprite{
			Sheet: uiSheet,
			Name:  gamerSprite,
			Scale: gs.Max,
		},
		geometry.Point{
			X: dr.Width * gs.Point.X * 0.5,
			Y: dr.Height * gs.Point.Y * 0.5,
		},
	)

	var size geometry.Size
	if size, err = eng.GetSpriteSize(uiSheet, logoSprite); err != nil {
		return err
	}

	// add logo
	world.AddEntity(
		sprite.Sprite{
			Sheet: uiSheet,
			Name:  logoSprite,
			Scale: gs.Max * logoScale,
		},
		geometry.Point{
			X: dr.Width * gs.Point.X * 0.5,
			Y: size.Height * gs.Point.Y * 0.5 * logoScale,
		},
	)

	// measuring the biggest text for size all the buttons equally
	var measure geometry.Size
	if measure, err = eng.MeasureText(font, " Options ", fontBigSize); err != nil {
		return err
	}

	measure.Width += measure.Width * buttonExtraWidth
	measure.Height += measure.Height * buttonExtraHeight

	buttonPos := geometry.Point{
		X: (dr.Width * gs.Point.X * 0.5) - (measure.Width * gs.Max * 0.5),
		Y: (dr.Height * gs.Point.Y) - (measure.Height * gs.Max * 2) - (20 * gs.Max),
	}

	// add the play button, it will sent a event to change to the main stage
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * gs.Max, Height: shadowExtraHeight * gs.Max},
			Event: events.DelaySignal{
				Signal: events.ChangeGameStage{Stage: "game"},
				Time:   0.25,
			},
			Sound:  clickSound,
			Volume: 1,
		},
		buttonPos,
		shapes.Box{
			Size: geometry.Size{
				Width:  measure.Width,
				Height: measure.Height,
			},
			Scale:     gs.Max,
			Thickness: int32(2 * gs.Max),
		},
		ui.Text{
			String:     "Play!",
			Size:       fontBigSize * gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		ui.ButtonColor{
			Gradient: color.Gradient{
				From: color.Red,
				To:   color.DarkPurple,
			},
			Border: color.White,
			Text:   color.SkyBlue,
		},
	)

	smallSize := geometry.Size{
		Width:  measure.Width * 0.48,
		Height: measure.Height * 0.75,
	}

	buttonPos = geometry.Point{
		X: (dr.Width * gs.Point.X * 0.5) - (measure.Width * 0.5 * gs.Max),
		Y: (dr.Height * gs.Point.Y) - (measure.Height * gs.Max) - (10 * gs.Max),
	}

	// add the options button, it will sent a event to change to the main stage
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * gs.Max, Height: shadowExtraHeight * gs.Max},
			Event: events.DelaySignal{
				Signal: events.ChangeGameStage{Stage: "game"},
				Time:   0.25,
			},
			Sound:  clickSound,
			Volume: 1,
		},
		buttonPos,
		shapes.Box{
			Size:      smallSize,
			Scale:     gs.Max,
			Thickness: int32(2 * gs.Max),
		},
		ui.Text{
			String:     "Options",
			Size:       fontSmallSize * gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		ui.ButtonColor{
			Gradient: color.Gradient{
				From: color.Red,
				To:   color.DarkPurple,
			},
			Border: color.White,
			Text:   color.SkyBlue,
		},
	)

	buttonPos = geometry.Point{
		X: (dr.Width * gs.Point.X * 0.5) + (((measure.Width * 0.5) - smallSize.Width) * gs.Max),
		Y: (dr.Height * gs.Point.Y) - (measure.Height * gs.Max) - (10 * gs.Max),
	}

	// add the exit button, it will sent a event to change to the main stage
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * gs.Max, Height: shadowExtraHeight * gs.Max},
			Event: events.DelaySignal{
				Signal: events.GameCloseEvent{},
				Time:   0.25,
			},
			Sound:  clickSound,
			Volume: 1,
		},
		buttonPos,
		shapes.Box{
			Size:      smallSize,
			Scale:     gs.Max,
			Thickness: int32(2 * gs.Max),
		},
		ui.Text{
			String:     "Exit",
			Size:       fontSmallSize * gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		ui.ButtonColor{
			Gradient: color.Gradient{
				From: color.Red,
				To:   color.DarkPurple,
			},
			Border: color.White,
			Text:   color.SkyBlue,
		},
	)

	world.AddListener(menuKeyListener)
	return world.Signal(events.PlayMusicEvent{Name: music, Volume: 1})
}

func menuKeyListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case events.KeyUpEvent:
		if e.Key == device.KeyReturn || e.Key == device.KeySpace {
			return world.Signal(events.ChangeGameStage{Stage: "game"})
		}
		break
	}
	return nil
}
