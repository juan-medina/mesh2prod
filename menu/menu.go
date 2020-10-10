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
	"fmt"
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/device"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/shapes"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/components/ui"
	"github.com/juan-medina/gosge/events"
	"reflect"
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
	mainMenu          = "main"                           // main menu
	optionsMenu       = "options"                        // options menu
	menuControlBorder = 2                                // menu controls border thickness
	music             = "resources/music/menu/Of Far Different Nature - Adventure Begins (CC-BY).ogg"
)

var (
	gEng        *gosge.Engine
	barEnt      *goecs.Entity
	valueLabel  *goecs.Entity
	currentMenu = mainMenu
)

// Stage the menu
func Stage(eng *gosge.Engine) error {
	var err error
	gEng = eng
	eng.DisableExitKey()

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

	// create the background
	if err = createBackGround(eng, world, dr, gs); err != nil {
		return err
	}

	// create main menu
	if err = createMainMenu(eng, world, dr, gs); err != nil {
		return err
	}

	// create options menu
	if err = createOptionsMenu(eng, world, dr, gs); err != nil {
		return err
	}

	world.AddListener(changeMenuListener, changeMenuEventType)

	// set the master volume to it config value
	currentMaster := eng.GetSettings().GetFloat32("master_volume", 1)

	// set the master volume
	world.Signal(events.ChangeMasterVolumeEvent{Volume: currentMaster})

	world.Signal(events.PlayMusicEvent{Name: music, Volume: 1})

	world.Signal(changeMenuEvent{name: mainMenu})

	return nil
}

func changeMenuListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case changeMenuEvent:
		currentMenu = e.name
		for it := world.Iterator(menuType); it != nil; it = it.Next() {
			ent := it.Value()
			mn := ent.Get(menuType).(menu)
			// show menu
			if mn.name == e.name {
				if mn.focus {
					world.Signal(events.FocusOnControlEvent{Control: ent})
				}
				if ent.Contains(effects.TYPE.Hide) {
					ent.Remove(effects.TYPE.Hide)
				}
			} else {
				if ent.NotContains(effects.TYPE.Hide) {
					ent.Add(effects.Hide{})
				}
			}
		}
	}
	return nil
}

func createBackGround(eng *gosge.Engine, world *goecs.World, dr geometry.Size, gs geometry.Scale) error {
	var err error
	var size geometry.Size
	if size, err = eng.GetSpriteSize(uiSheet, logoSprite); err != nil {
		return err
	}

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
	return nil
}

func createMainMenu(eng *gosge.Engine, world *goecs.World, dr geometry.Size, gs geometry.Scale) error {
	var err error

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
			Thickness: int32(menuControlBorder * gs.Max),
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
			Border: color.DarkBlue,
			Text:   color.SkyBlue,
		},
		menu{name: mainMenu, focus: true},
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
				Signal: changeMenuEvent{name: optionsMenu},
				Time:   0.25,
			},
			Sound:  clickSound,
			Volume: 1,
		},
		buttonPos,
		shapes.Box{
			Size:      smallSize,
			Scale:     gs.Max,
			Thickness: int32(menuControlBorder * gs.Max),
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
			Border: color.DarkBlue,
			Text:   color.SkyBlue,
		},
		menu{name: mainMenu},
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
			Thickness: int32(menuControlBorder * gs.Max),
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
			Border: color.DarkBlue,
			Text:   color.SkyBlue,
		},
		menu{name: mainMenu},
	)

	world.AddListener(menuKeyListener, events.TYPE.KeyUpEvent)

	world.AddListener(gamepadListener, events.TYPE.GamePadButtonUpEvent)

	return nil
}

func createOptionsMenu(eng *gosge.Engine, world *goecs.World, dr geometry.Size, gs geometry.Scale) error {
	currentMaster := float32(int(eng.GetSettings().GetFloat32("master_volume", 1) * 100))

	panelSize := geometry.Size{
		Width:  520,
		Height: 210,
	}

	panelPos := geometry.Point{
		X: (dr.Width * gs.Point.X * 0.5) - (panelSize.Width * gs.Max * 0.5),
		Y: (dr.Height * gs.Point.Y * 0.5) - (panelSize.Height * gs.Max * 0.5),
	}
	world.AddEntity(
		shapes.SolidBox{
			Size:  panelSize,
			Scale: gs.Max,
		},
		panelPos,
		color.Black.Alpha(90),
		menu{name: optionsMenu},
		effects.Hide{},
	)
	world.AddEntity(
		shapes.Box{
			Size:      panelSize,
			Scale:     gs.Max,
			Thickness: int32(menuControlBorder * gs.Max),
		},
		panelPos,
		color.White,
		menu{name: optionsMenu},
		effects.Hide{},
	)

	labelPos := geometry.Point{
		X: panelPos.X + (panelSize.Width * 0.5 * gs.Max),
		Y: panelPos.Y + (40 * gs.Max),
	}

	world.AddEntity(
		ui.Text{
			String:     "Options",
			Size:       fontBigSize * gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		labelPos,
		color.White,
		menu{name: optionsMenu},
		effects.Hide{},
	)

	labelPos = geometry.Point{
		X: panelPos.X + (10 * gs.Max),
		Y: panelPos.Y + (90 * gs.Max),
	}

	world.AddEntity(
		ui.Text{
			String:     "Master Volume",
			Size:       fontSmallSize * gs.Max,
			Font:       font,
			VAlignment: ui.TopVAlignment,
			HAlignment: ui.LeftHAlignment,
		},
		labelPos,
		color.White,
		menu{name: optionsMenu},
		effects.Hide{},
	)

	controlPos := geometry.Point{
		X: labelPos.X + (200 * gs.Max),
		Y: labelPos.Y,
	}

	controlSize := geometry.Size{
		Width:  300,
		Height: 40,
	}

	bar := ui.ProgressBar{
		Min:     0,
		Max:     100,
		Current: currentMaster,
		Shadow: geometry.Size{
			Width:  2 * gs.Max,
			Height: 2 * gs.Max,
		},
		Sound:  clickSound,
		Volume: 1,
		Event:  masterVolumeChangeEvent{},
	}

	// add the master volume progress bar
	barEnt = world.AddEntity(
		ui.ProgressBarColor{
			Solid: color.SkyBlue,
			Gradient: color.Gradient{
				From:      color.SkyBlue,
				To:        color.DarkBlue,
				Direction: color.GradientHorizontal,
			},
			Empty:  color.Blue.Blend(color.White, 0.65),
			Border: color.DarkBlue,
		},
		shapes.Box{
			Size:      controlSize,
			Scale:     gs.Max,
			Thickness: int32(menuControlBorder * gs.Max),
		},
		controlPos,
		menu{name: optionsMenu},
		effects.Hide{},
	)

	labelPos = geometry.Point{
		X: controlPos.X + (controlSize.Width * 0.5 * gs.Max),
		Y: controlPos.Y + (controlSize.Height * 0.5 * gs.Max),
	}

	text := "Muted"
	if currentMaster != 0 {
		text = fmt.Sprintf("%d%%", int(currentMaster))
	}

	valueLabel = world.AddEntity(
		ui.Text{
			String:     text,
			Size:       fontSmallSize * gs.Max,
			Font:       font,
			VAlignment: ui.MiddleVAlignment,
			HAlignment: ui.CenterHAlignment,
		},
		labelPos,
		color.White,
		menu{name: optionsMenu},
		effects.Hide{},
	)

	barEnt.Set(bar)

	controlSize = geometry.Size{
		Width:  100,
		Height: 40,
	}

	controlPos.Y += 60 * gs.Max

	// add the save button
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * gs.Max, Height: shadowExtraHeight * gs.Max},
			Event:  saveOptionsEvent{},
			Sound:  clickSound,
			Volume: 1,
		},
		controlPos,
		shapes.Box{
			Size:      controlSize,
			Scale:     gs.Max,
			Thickness: int32(menuControlBorder * gs.Max),
		},
		ui.Text{
			String:     "Save",
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
			Border: color.DarkBlue,
			Text:   color.SkyBlue,
		},
		menu{name: optionsMenu, focus: true},
		effects.Hide{},
	)

	controlPos.X = controlPos.X + ((controlSize.Width + 10) * gs.Max)

	// add the cancel button
	world.AddEntity(
		ui.FlatButton{
			Shadow: geometry.Size{Width: shadowExtraWidth * gs.Max, Height: shadowExtraHeight * gs.Max},
			Event:  cancelOptionsEvent{},
			Sound:  clickSound,
			Volume: 1,
		},
		controlPos,
		shapes.Box{
			Size:      controlSize,
			Scale:     gs.Max,
			Thickness: int32(menuControlBorder * gs.Max),
		},
		ui.Text{
			String:     "Cancel",
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
			Border: color.DarkBlue,
			Text:   color.SkyBlue,
		},
		menu{name: optionsMenu},
		effects.Hide{},
	)

	world.AddListener(optionsListener, masterVolumeChangeEventType, cancelOptionsEventType, saveOptionsEventType)

	return nil
}

func optionsListener(world *goecs.World, signal interface{}, _ float32) error {
	switch signal.(type) {
	case masterVolumeChangeEvent:
		bar := ui.Get.ProgressBar(barEnt)
		text := ui.Get.Text(valueLabel)
		value := int(bar.Current)
		if value != 0 {
			text.String = fmt.Sprintf("%d%%", value)
		} else {
			text.String = "Muted"
		}
		valueLabel.Set(text)
		world.Signal(events.ChangeMasterVolumeEvent{Volume: float32(value) / 100})
	case cancelOptionsEvent:
		master := float32(int(gEng.GetSettings().GetFloat32("master_volume", 1) * 100))
		bar := ui.Get.ProgressBar(barEnt)
		bar.Current = master
		barEnt.Set(bar)
		text := ui.Get.Text(valueLabel)
		text.String = fmt.Sprintf("%d%%", int32(master))
		valueLabel.Set(text)
		world.Signal(events.ChangeMasterVolumeEvent{Volume: master / 100})
		world.Signal(events.DelaySignal{
			Signal: changeMenuEvent{name: mainMenu},
			Time:   0.25,
		})
	case saveOptionsEvent:
		bar := ui.Get.ProgressBar(barEnt)
		value := bar.Current / 100
		gEng.GetSettings().SetFloat32("master_volume", value)
		world.Signal(events.DelaySignal{
			Signal: changeMenuEvent{name: mainMenu},
			Time:   0.25,
		})
	}
	return nil
}

func menuKeyListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case events.KeyUpEvent:
		if e.Key == device.KeyReturn {
			switch currentMenu {
			case mainMenu:
				world.Signal(events.ChangeGameStage{Stage: "game"})
			case optionsMenu:
				world.Signal(saveOptionsEvent{})
			}
		} else if e.Key == device.KeyEscape {
			switch currentMenu {
			case mainMenu:
				world.Signal(events.GameCloseEvent{})
			case optionsMenu:
				world.Signal(cancelOptionsEvent{})
			}
		}
	}
	return nil
}

func gamepadListener(world *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case events.GamePadButtonUpEvent:
		if e.Button == device.GamepadStart {
			switch currentMenu {
			case mainMenu:
				world.Signal(events.ChangeGameStage{Stage: "game"})
			case optionsMenu:
				world.Signal(saveOptionsEvent{})
			}
		} else if e.Button == device.GamepadSelect || e.Button == device.GamepadButton2 {
			switch currentMenu {
			case mainMenu:
				world.Signal(events.GameCloseEvent{})
			case optionsMenu:
				world.Signal(cancelOptionsEvent{})
			}
		}
	}
	return nil
}

type masterVolumeChangeEvent struct{}

var masterVolumeChangeEventType = reflect.TypeOf(masterVolumeChangeEvent{})

type cancelOptionsEvent struct{}

var cancelOptionsEventType = reflect.TypeOf(cancelOptionsEvent{})

type saveOptionsEvent struct{}

var saveOptionsEventType = reflect.TypeOf(saveOptionsEvent{})

type menu struct {
	name  string
	focus bool
}

var menuType = reflect.TypeOf(menu{})

type changeMenuEvent struct {
	name string
}

var changeMenuEventType = reflect.TypeOf(changeMenuEvent{})
