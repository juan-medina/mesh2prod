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

package constants

// game constants
const (
	SpriteSheet         = "resources/sprites/mesh2prod.json" // game sprite sheet
	CloudSizeConfig     = "cloud_size"                       // cloud side config value
	MasterVolumeConfig  = "master_volume"                    // master volume config setting
	DefaultMasterVolume = 1                                  // Default master volume
)

// CloudSize is the cloud size
type CloudSize int

// clouds
const (
	LocalCloud   = CloudSize(iota) // LocalCloudSize cloud size
	StartupCloud                   // StartupCloudSize cloud size
	CorpCloud                      // CorpCloudSize cloud size
	PublicCloud                    // PublicCloudSize cloud size
)

// clouds
var (
	// Clouds is our clouds types
	Clouds = []CloudSize{LocalCloud, StartupCloud, CorpCloud, PublicCloud}

	// CloudNames is our cloud names
	CloudNames = map[CloudSize]string{
		LocalCloud:   "local",
		StartupCloud: "startup",
		CorpCloud:    "corp",
		PublicCloud:  "public",
	}

	// CloudSizes is our cloud sizes
	CloudSizes = map[CloudSize]int{
		LocalCloud:   50,
		StartupCloud: 150,
		CorpCloud:    300,
		PublicCloud:  450,
	}
)
