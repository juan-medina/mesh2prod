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

package main

import (
	"os"
	"path/filepath"

	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/gosge/options"
	"github.com/juan-medina/mesh2prod/game"
	"github.com/juan-medina/mesh2prod/intro"
	"github.com/juan-medina/mesh2prod/menu"
	"github.com/rs/zerolog/log"
)

// general constants
const (
	version = "mesh2prod : 1.0.0.alpha"
)

// game options
var opt = options.Options{
	Title:      "mesh2prod",
	BackGround: color.Black,
	Icon:       "resources/icon/icon.png",
	// Uncomment this for using windowed mode
	// Windowed: true,
	// Width:    2048,
	// Height:   1536,
}

func load(eng *gosge.Engine) error {
	eng.GetSettings().SetString("version", version)
	eng.AddGameStage("game", game.Stage)
	eng.AddGameStage("menu", menu.Stage)
	eng.AddGameStage("intro", intro.Stage)
	eng.World().Signal(events.ChangeGameStage{Stage: "intro"})
	return nil
}

func main() {
	var err error = nil
	// if we can not find the resources folder
	if _, err = os.Stat("resources"); os.IsNotExist(err) {
		var dir string
		// get our executable path
		if dir, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			log.Fatal().Err(err).Msg("error checking path")
		}
		// change working directory
		if err = os.Chdir(dir); err != nil {
			log.Fatal().Err(err).Msg("error changing path")
		}
		// final check if now we could find the resources
		if _, err = os.Stat("resources"); os.IsNotExist(err) {
			log.Fatal().Msg("can't find resources")
		}
	}
	// run the game
	if err = gosge.Run(opt, load); err != nil {
		log.Fatal().Err(err).Msg("error running the game")
	}
}
