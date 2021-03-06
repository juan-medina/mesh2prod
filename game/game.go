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
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/background"
	"github.com/juan-medina/mesh2prod/game/collision"
	"github.com/juan-medina/mesh2prod/game/constants"
	"github.com/juan-medina/mesh2prod/game/gamemap"
	"github.com/juan-medina/mesh2prod/game/mesh"
	"github.com/juan-medina/mesh2prod/game/movement"
	"github.com/juan-medina/mesh2prod/game/music"
	"github.com/juan-medina/mesh2prod/game/plane"
	"github.com/juan-medina/mesh2prod/game/score"
	"github.com/juan-medina/mesh2prod/game/target"
	"github.com/juan-medina/mesh2prod/game/winning"
	"math/rand"
	"time"
)

const (
	planeLoop = "resources/audio/plane.ogg" // our plane loop
)

var (
	designResolution = geometry.Size{Width: 1920, Height: 1080} // designResolution is how our game is designed
)

// Stage the game
func Stage(eng *gosge.Engine) error {
	var err error
	rand.Seed(time.Now().UnixNano())
	eng.DisableExitKey()

	// get the ECS world
	world := eng.World()

	// gameScale from the real screen size to our design resolution
	gameScale := eng.GetScreenSize().CalculateScale(designResolution)

	// get random music
	var musicFile string
	if musicFile, err = music.GetRandomMusic(); err != nil {
		return err
	}

	// load the music
	if err = eng.LoadMusic(musicFile); err != nil {
		return err
	}

	// load the plane loop
	if err = eng.LoadMusic(planeLoop); err != nil {
		return err
	}

	// load the sprite sheet
	if err = eng.LoadSpriteSheet(constants.SpriteSheet); err != nil {
		return err
	}

	// add movement system
	if err = movement.System(eng, gameScale); err != nil {
		return err
	}

	// add the plane
	if err = plane.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	// add the background system
	if err = background.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	// add the mesh
	if err = mesh.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	cs := constants.CloudSize(eng.GetSettings().GetIn32(constants.CloudSizeConfig, int32(constants.StartupCloud)))
	length := constants.CloudSizes[cs]

	// add the map
	if err = gamemap.System(eng, gameScale, designResolution, int(length)); err != nil {
		return err
	}

	// add the target
	if err = target.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	// add the collision system
	if err = collision.System(eng); err != nil {
		return err
	}

	// add the score system
	if err = score.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	// add the winning system
	if err = winning.System(eng, gameScale, designResolution); err != nil {
		return err
	}

	// play the music
	world.Signal(events.PlayMusicEvent{Name: musicFile, Volume: 0.5})

	// play the plane loop
	world.Signal(events.PlayMusicEvent{Name: planeLoop, Volume: 1})

	return nil
}
