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

package component

import (
	"github.com/juan-medina/goecs"
	"reflect"
)

// Type represents a Tag type
type Type int

// Tag Types
//goland:noinspection GoUnusedConst
const (
	None       = Type(iota) // None tag
	BulletType              // Bullet tag
)

// Bullet is a component for our bullets
type Bullet struct{}

type types struct {
	// Bullet is the reflect.Type for component.Bullet
	Bullet reflect.Type
}

// TYPE hold the reflect.Type for our components
var TYPE = types{
	Bullet: reflect.TypeOf(Bullet{}),
}

type gets struct {
	// Bullet gets a component.Bullet from a goecs.Entity
	Bullet func(e *goecs.Entity) Bullet
}

// Get a geometry component
//goland:noinspection GoUnusedGlobalVariable
var Get = gets{
	// Point gets a component.Bullet from a goecs.Entity
	Bullet: func(e *goecs.Entity) Bullet {
		return e.Get(TYPE.Bullet).(Bullet)
	},
}
