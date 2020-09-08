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
	"github.com/juan-medina/gosge/components/geometry"
	"reflect"
)

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

// indicate how much we need to move
type movement struct {
	amount geometry.Point // how much we could move
	min    geometry.Point // min position that we could move
	max    geometry.Point // max position that we could move
}

var movementType = reflect.TypeOf(movement{})
