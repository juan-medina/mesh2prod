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

package winning

import (
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/component"
)

const (
	font              = "resources/fonts/go_regular.fnt" // our message text font
	fontSize          = 60                               // message text font size
	fontSmall         = 40                               //  message text small font size
	fontButtonSize    = 30                               // font size for button
	shadowExtraWidth  = 3                                // the x offset for the buttons shadow
	shadowExtraHeight = 3                                // the y offset for the buttons shadow
	buttonExtraWidth  = 0.15                             // the additional width for a button si it is not only the text size
	buttonExtraHeight = 0.20                             // the additional width for a button si it is not only the text size
	clickSound        = "resources/audio/click.wav"      // button click sound
	winSound          = "resources/audio/win.wav"        // win sound
	music             = "resources/music/loop.ogg"       // our game music
)

// FinalScoreEvent is trigger when the game ends
type FinalScoreEvent struct {
	Total int
}

var (
	bcColor = color.Solid{R: 227, G: 140, B: 41, A: 255} // our bc text color
)

// LevelEndEvent is trigger when the level end
type LevelEndEvent struct{}

type winningSystem struct {
	gs       geometry.Scale
	dr       geometry.Size
	eng      *gosge.Engine
	end      bool
	label    *goecs.Entity
	sprint   *goecs.Entity
	distance float32
}

// add the background
func (ws *winningSystem) load(eng *gosge.Engine) error {
	var err error

	// pre-load font
	if err = eng.LoadFont(font); err != nil {
		return err
	}

	// pre-load click sound
	if err = eng.LoadSound(clickSound); err != nil {
		return err
	}

	// pre-load win sound
	if err = eng.LoadSound(winSound); err != nil {
		return err
	}

	// get the ECS world
	world := eng.World()

	sprintPos := geometry.Point{
		X: 5 * ws.gs.Max,
		Y: 5 * ws.gs.Max,
	}

	ws.sprint = world.AddEntity(
		ui.Text{
			String:     "Sprinting Points to prod : 0",
			Size:       fontSmall * ws.gs.Max,
			Font:       font,
			VAlignment: ui.TopVAlignment,
			HAlignment: ui.LeftHAlignment,
		},
		sprintPos,
		color.White,
	)

	// calculate when we reach production
	world.AddSystem(ws.reachProductionSystem)

	// final score listener
	world.AddListener(ws.finalScoreListener)

	// update sprint system
	world.AddSystem(ws.updateSprintSystem)

	// listen to keys
	world.AddListener(ws.KeysListener)

	return nil
}

func (ws *winningSystem) reachProductionSystem(world *goecs.World, _ float32) error {
	if ws.end {
		return nil
	}
	mesh := world.Iterator(component.TYPE.Mesh).Value()
	prod := world.Iterator(component.TYPE.Production).Value()

	meshPos := geometry.Get.Point(mesh)
	prodPos := geometry.Get.Point(prod)

	diffX := prodPos.X - meshPos.X
	if diffX < 0 {
		ws.end = true
		_ = world.Signal(LevelEndEvent{})
		_ = world.Signal(events.StopMusicEvent{Name: music})
	}

	return nil
}

func (ws *winningSystem) addMessage(world *goecs.World) error {
	boxSize := geometry.Size{
		Width:  ws.dr.Width * 0.35,
		Height: ws.dr.Height * 0.25,
	}

	boxPos := geometry.Point{
		X: ((ws.dr.Width * ws.gs.Point.X) - (boxSize.Width * ws.gs.Max)) * 0.5,
		Y: ((ws.dr.Height * ws.gs.Point.Y) - (boxSize.Height * ws.gs.Max)) * 0.5,
	}

	textPos := geometry.Point{
		X: boxPos.X + (boxSize.Width * ws.gs.Max * 0.5),
		Y: boxPos.Y + (10 * ws.gs.Max),
	}

	world.AddEntity(
		shapes.SolidBox{
			Size:  boxSize,
			Scale: ws.gs.Max,
		},
		color.Gradient{
			From:      color.DarkBlue.Alpha(210),
			To:        color.SkyBlue.Alpha(190),
			Direction: color.GradientVertical,
		},
		boxPos,
		effects.Layer{Depth: -2},
	)
	world.AddEntity(
		shapes.Box{
			Size:      boxSize,
			Scale:     ws.gs.Max,
			Thickness: int32(2 * ws.gs.Max),
		},
		color.DarkBlue,
		boxPos,
		effects.Layer{Depth: -2},
	)

	world.AddEntity(
		ui.Text{
			String:     "Delivered to Prod!",
			Size:       fontSize * ws.gs.Max,
			Font:       font,
			VAlignment: ui.TopVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		textPos,
		color.White,
		effects.Layer{Depth: -2},
	)

	textPos.Y = boxPos.Y + (boxSize.Height * ws.gs.Point.Y * 0.5)

	ws.label = world.AddEntity(
		ui.Text{
			String:     "You got 0 BlockCoins",
			Size:       fontSmall * ws.gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		textPos,
		bcColor,
		effects.Layer{Depth: -2},
	)

	// measuring the biggest text for size all the buttons equally
	var measure geometry.Size
	var err error
	if measure, err = ws.eng.MeasureText(font, " Yay ! ", fontButtonSize); err != nil {
		return err
	}

	measure.Width += measure.Width * buttonExtraWidth
	measure.Height += measure.Height * buttonExtraHeight

	buttonPos := geometry.Point{
		X: textPos.X - (measure.Width * ws.gs.Max * 0.5),
		Y: boxPos.Y + (boxSize.Height * ws.gs.Max) - (measure.Height * ws.gs.Max * 1.25),
	}

	// add the play button, it will sent a event to change to the main stage
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * ws.gs.Max, Height: shadowExtraHeight * ws.gs.Max},
			Event: events.DelaySignal{
				Signal: events.ChangeGameStage{Stage: "menu"},
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
			Scale:     ws.gs.Max,
			Thickness: int32(2 * ws.gs.Max),
		},
		ui.Text{
			String:     "Yay!",
			Size:       fontButtonSize * ws.gs.Max,
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
		effects.Layer{Depth: -2},
	)
	return nil
}

func (ws *winningSystem) finalScoreListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case FinalScoreEvent:
		if err := ws.addMessage(world); err != nil {
			return err
		}
		text := ui.Get.Text(ws.label)
		text.String = fmt.Sprintf("You got %d BlockCoins", e.Total)
		ws.label.Set(text)
		return world.Signal(events.PlaySoundEvent{Name: winSound, Volume: 1})
	}
	return nil
}

func (ws *winningSystem) KeysListener(world *goecs.World, signal interface{}, _ float32) error {
	if !ws.end {
		return nil
	}
	switch e := signal.(type) {
	case events.KeyUpEvent:
		if e.Key == device.KeySpace {
			_ = world.Signal(events.PlaySoundEvent{Name: clickSound, Volume: 1})
			return world.Signal(events.DelaySignal{
				Signal: events.ChangeGameStage{Stage: "game"},
				Time:   0.25,
			})
		}
	}
	return nil
}

func (ws *winningSystem) updateSprintSystem(world *goecs.World, _ float32) error {
	if ws.end {
		return nil
	}

	mesh := world.Iterator(component.TYPE.Mesh).Value()
	prod := world.Iterator(component.TYPE.Production).Value()

	meshPos := geometry.Get.Point(mesh)
	prodPos := geometry.Get.Point(prod)

	diff := prodPos.X - meshPos.X

	if ws.distance == 0 {
		ws.distance = diff
	}

	percent := diff / ws.distance

	sprints := int(percent * 1024)

	text := ui.Get.Text(ws.sprint)

	text.String = fmt.Sprintf("Sprinting Points to prod : %d", sprints)
	ws.sprint.Set(text)

	return nil
}

// System creates the mesh system
func System(engine *gosge.Engine, gs geometry.Scale, dr geometry.Size) error {
	ws := winningSystem{
		gs:  gs,
		dr:  dr,
		eng: engine,
	}
	return ws.load(engine)
}
