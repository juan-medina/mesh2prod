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
	"github.com/juan-medina/gosge/components/color"
	"github.com/juan-medina/gosge/components/effects"
	"github.com/juan-medina/gosge/components/geometry"
	"github.com/juan-medina/gosge/components/sprite"
	"github.com/juan-medina/gosge/events"
	"github.com/juan-medina/mesh2prod/game/component"
)

type collisionSystem struct {
	eng *gosge.Engine
}

func (cs *collisionSystem) load(engine *gosge.Engine) error {
	// get the world
	world := engine.World()

	// add the bullet <-> block collision system
	world.AddSystem(cs.blocksCollisionsSystem)

	world.AddListener(cs.removeTintsListener)
	return nil
}

func (cs *collisionSystem) blocksCollisionsSystem(world *goecs.World, _ float32) error {
	for it := world.Iterator(geometry.TYPE.Point, sprite.TYPE); it != nil; it = it.Next() {
		ent := it.Value()
		if ent.Contains(component.TYPE.Bullet) {
			block := cs.checkBlocks(ent, world)
			if block != nil {
				blockC := component.Get.Block(block)
				if err := world.Signal(BulletHitBlockEvent{Block: blockC}); err != nil {
					return err
				}
				_ = world.Remove(ent)
				continue
			}
		} else if ent.Contains(component.TYPE.Plane) {
			if cs.checkPlaneBlock(ent, world) {
				cs.tintEntity(ent, world)
			}
		} else if ent.Contains(component.TYPE.Mesh) {
			if cs.checkMeshBlock(ent, world) {
				cs.tintEntity(ent, world)
			}
		}
	}

	return nil
}
func (cs *collisionSystem) checkBlocks(bullet *goecs.Entity, world *goecs.World) *goecs.Entity {
	for it := world.Iterator(component.TYPE.Block, geometry.TYPE.Point, sprite.TYPE); it != nil; it = it.Next() {
		block := it.Value()
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

func (cs *collisionSystem) checkPlaneBlock(plane *goecs.Entity, world *goecs.World) bool {
	any := false
	for it := world.Iterator(component.TYPE.Block, geometry.TYPE.Point, sprite.TYPE); it != nil; it = it.Next() {
		block := it.Value()
		if cs.spriteCollide(plane, block) {
			any = true
			cs.removeBlock(block, world, true)
		}
	}
	return any
}

func (cs *collisionSystem) checkMeshBlock(mesh *goecs.Entity, world *goecs.World) bool {
	any := false
	for it := world.Iterator(component.TYPE.Block, geometry.TYPE.Point, sprite.TYPE); it != nil; it = it.Next() {
		block := it.Value()
		if cs.spriteCollide(mesh, block) {
			any = true
			cs.removeBlock(block, world, false)
		}
	}
	return any
}

func (cs *collisionSystem) removeBlock(block *goecs.Entity, world *goecs.World, isPlane bool) {
	blockC := component.Get.Block(block)
	if blockC.Text != nil {
		_ = world.Remove(blockC.Text)
		blockC.Text = nil
		block.Set(blockC)
	}

	_ = world.Remove(block)
	if isPlane {
		_ = world.Signal(PlaneHitBlockEvent{Block: blockC})
	} else {
		_ = world.Signal(MeshHitBlockEvent{Block: blockC})
	}
}

func (cs *collisionSystem) tintEntity(ent *goecs.Entity, world *goecs.World) {
	if ent.NotContains(effects.TYPE.AlternateColor) {
		ent.Add(effects.AlternateColor{
			From:  color.White,
			To:    color.Red,
			Time:  0.15,
			Delay: 0,
		})
		_ = world.Signal(events.DelaySignal{
			Signal: RemoveTintEvent{ent: ent},
			Time:   1.5,
		})
	}
}

func (cs *collisionSystem) removeTintsListener(_ *goecs.World, signal interface{}, _ float32) error {
	switch e := signal.(type) {
	case RemoveTintEvent:
		e.ent.Remove(effects.TYPE.AlternateColor)
		e.ent.Remove(effects.TYPE.AlternateColorState)
		e.ent.Set(color.White)
	}
	return nil
}

// BulletHitBlockEvent is trigger when a bullet hit a block
type BulletHitBlockEvent struct {
	Block component.Block
}

// PlaneHitBlockEvent is trigger when the plane hit a block
type PlaneHitBlockEvent struct {
	Block component.Block
}

// MeshHitBlockEvent is trigger when the mesh hit a block
type MeshHitBlockEvent struct {
	Block component.Block
}

// RemoveTintEvent is a event to remove a tint
type RemoveTintEvent struct {
	ent *goecs.Entity
}

// System create the map system
func System(engine *gosge.Engine) error {
	cs := collisionSystem{
		eng: engine,
	}

	return cs.load(engine)
}
