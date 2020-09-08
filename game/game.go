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
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/color"
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

	// add the background
	if err = addBackground(eng); err != nil {
		return err
	}

	// add the mesh
	if err = addMesh(eng); err != nil {
		return err
	}

	// add the plane
	if err = addPlane(eng); err != nil {
		return err
	}

	// add the move system
	world.AddSystem(moveSystem)

	// play the music
	return world.Signal(events.PlayMusicEvent{Name: music})
}
