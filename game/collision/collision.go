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

package collision

import (
	"github.com/juan-medina/goecs"
	"github.com/juan-medina/gosge"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/mesh2prod/game/component"
)

type collisionSystem struct {
	eng *gosge.Engine
}

func (cs *collisionSystem) load(engine *gosge.Engine) error {
	// get the world
	world := engine.World()

	// add the bullet <-> block collision system
	world.AddSystem(cs.bulletBlocksCollisionsSystem)
	return nil
}

func (cs *collisionSystem) bulletBlocksCollisionsSystem(world *goecs.World, _ float32) error {

	for itBu := world.Iterator(component.TYPE.Bullet, geometry.TYPE.Point, sprite.TYPE); itBu != nil; itBu = itBu.Next() {
		bullet := itBu.Value()
		block := cs.checkBlocks(bullet, world)
		if block != nil {
			blockC := component.Get.Block(block)
			if err := world.Signal(BulletHitBlockEvent{Block: blockC}); err != nil {
				return err
			}
			_ = world.Remove(bullet)
			continue
		}
	}

	return nil
}
func (cs *collisionSystem) checkBlocks(bullet *goecs.Entity, world *goecs.World) *goecs.Entity {
	for itBl := world.Iterator(component.TYPE.Block, geometry.TYPE.Point, sprite.TYPE); itBl != nil; itBl = itBl.Next() {
		block := itBl.Value()
		if cs.spriteCollide(bullet, block) {
			return block
		}
	}
	return nil
}

func (cs *collisionSystem) spriteCollide(ent1, ent2 *goecs.Entity) bool {
	spr1 := sprite.Get(ent1)
	pos1 := geometry.Get.Point(ent1)

	spr2 := sprite.Get(ent2)
	pos2 := geometry.Get.Point(ent2)

	return cs.eng.SpritesCollides(spr1, pos1, spr2, pos2)
}

// BulletHitBlockEvent is trigger when a bullet hit a block
type BulletHitBlockEvent struct {
	Block component.Block
}

// System create the map system
func System(engine *gosge.Engine) error {
	cs := collisionSystem{
		eng: engine,
	}

	return cs.load(engine)
}
