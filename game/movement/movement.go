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

package movement

import (
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/geometry"
	"reflect"
)

type movementSystem struct {
	gs geometry.Scale
}

// move system
func (ms movementSystem) system(world *goecs.World, delta float32) error {
	// move anything that has a position and Movement
	for it := world.Iterator(geometry.TYPE.Point, Type); it != nil; it = it.Next() {
		// get the entity
		ent := it.Value()

		// get current position and Movement
		pos := geometry.Get.Point(ent)
		mov := ent.Get(Type).(Movement)

		// increment position and clamp to the Min/Max
		pos.Y += mov.Amount.Y * delta * ms.gs.Point.X
		pos.X += mov.Amount.X * delta * ms.gs.Point.Y

		// if we have constrains
		if ent.Contains(ConstrainType) {
			// clamp to them
			constrain := ent.Get(ConstrainType).(Constrain)
			pos.Clamp(constrain.Min, constrain.Max)
		}

		// update entity
		ent.Set(pos)
	}

	return nil
}

// Constrain of the movement
type Constrain struct {
	Min geometry.Point // Min position that we could move
	Max geometry.Point // Max position that we could move
}

// ConstrainType is the reflect.Type of Movement
var ConstrainType = reflect.TypeOf(Constrain{})

// Movement indicate how much we need to move
type Movement struct {
	Amount geometry.Point // Amount that we could move
}

// Type is the reflect.Type of Movement
var Type = reflect.TypeOf(Movement{})

// System Create the Movement system
func System(engine *gosge.Engine, gs geometry.Scale) error {
	ms := movementSystem{
		gs: gs,
	}

	engine.World().AddSystem(ms.system)

	return nil
}
