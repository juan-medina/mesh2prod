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

// Bullet is a component for our bullets
type Bullet struct{}

// Block is a component for a map blocks
type Block struct {
	C, R    int
	ClearOn float32
	Text    *goecs.Entity
}

// FloatText is a component for a floating text
type FloatText struct{}

// Plane is a component for the plane
type Plane struct{}

// Mesh is a component for the mesh
type Mesh struct{}

// Production is a component for the production area
type Production struct{}

type types struct {
	// Bullet is the reflect.Type for component.Bullet
	Bullet reflect.Type
	// Block is the reflect.Type for component.Block
	Block reflect.Type
	// FloatText is the reflect.Type for component.FloatText
	FloatText reflect.Type
	// Plane is the reflect.Type for component.Plane
	Plane reflect.Type
	// Mesh is the reflect.Type for component.Mesh
	Mesh reflect.Type
	// Production is the reflect.Type for component.Production
	Production reflect.Type
}

// TYPE hold the reflect.Type for our components
var TYPE = types{
	Bullet:     reflect.TypeOf(Bullet{}),
	Block:      reflect.TypeOf(Block{}),
	FloatText:  reflect.TypeOf(FloatText{}),
	Plane:      reflect.TypeOf(Plane{}),
	Mesh:       reflect.TypeOf(Mesh{}),
	Production: reflect.TypeOf(Production{}),
}

type gets struct {
	// Bullet gets a component.Bullet from a goecs.Entity
	Bullet func(e *goecs.Entity) Bullet
	// Block gets a component.Block from a goecs.Entity
	Block func(e *goecs.Entity) Block
	// FloatText gets a component.FloatText from a goecs.Entity
	FloatText func(e *goecs.Entity) FloatText
	// Plane gets a component.Plane from a goecs.Entity
	Plane func(e *goecs.Entity) Plane
	// Mesh gets a component.Mesh from a goecs.Entity
	Mesh func(e *goecs.Entity) Mesh
	// Production gets a component.Production from a goecs.Entity
	Production func(e *goecs.Entity) Production
}

// Get a geometry component
//goland:noinspection GoUnusedGlobalVariable
var Get = gets{
	// Bullet gets a component.Bullet from a goecs.Entity
	Bullet: func(e *goecs.Entity) Bullet {
		return e.Get(TYPE.Bullet).(Bullet)
	},
	// Bullet gets a component.Bullet from a goecs.Entity
	Block: func(e *goecs.Entity) Block {
		return e.Get(TYPE.Block).(Block)
	},
	// FloatText gets a component.FloatText from a goecs.Entity
	FloatText: func(e *goecs.Entity) FloatText {
		return e.Get(TYPE.FloatText).(FloatText)
	},
	// Plane gets a component.Plane from a goecs.Entity
	Plane: func(e *goecs.Entity) Plane {
		return e.Get(TYPE.Plane).(Plane)
	},
	// Mesh gets a component.Mesh from a goecs.Entity
	Mesh: func(e *goecs.Entity) Mesh {
		return e.Get(TYPE.Mesh).(Mesh)
	},
	// Production gets a component.Production from a goecs.Entity
	Production: func(e *goecs.Entity) Production {
		return e.Get(TYPE.Production).(Production)
	},
}
