// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package inpututil provides utility functions of input like keyboard or mouse.
//
// Note: This package is experimental and API might be changed.
package inpututil

import (
	"sort"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/internal/hooks"
	"github.com/hajimehoshi/ebiten/internal/sync"
)

type inputState struct {
	keyDurations     map[ebiten.Key]int
	prevKeyDurations map[ebiten.Key]int

	mouseButtonDurations     map[ebiten.MouseButton]int
	prevMouseButtonDurations map[ebiten.MouseButton]int

	gamepadIDs     map[int]struct{}
	prevGamepadIDs map[int]struct{}

	gamepadButtonDurations     map[int]map[ebiten.GamepadButton]int
	prevGamepadButtonDurations map[int]map[ebiten.GamepadButton]int

	touchDurations     map[int]int
	prevTouchDurations map[int]int

	m sync.RWMutex
}

var theInputState = &inputState{
	keyDurations:     map[ebiten.Key]int{},
	prevKeyDurations: map[ebiten.Key]int{},

	mouseButtonDurations:     map[ebiten.MouseButton]int{},
	prevMouseButtonDurations: map[ebiten.MouseButton]int{},

	gamepadIDs:     map[int]struct{}{},
	prevGamepadIDs: map[int]struct{}{},

	gamepadButtonDurations:     map[int]map[ebiten.GamepadButton]int{},
	prevGamepadButtonDurations: map[int]map[ebiten.GamepadButton]int{},

	touchDurations:     map[int]int{},
	prevTouchDurations: map[int]int{},
}

func init() {
	hooks.AppendHookOnBeforeUpdate(func() error {
		theInputState.update()
		return nil
	})
}

func (i *inputState) update() {
	i.m.Lock()
	defer i.m.Unlock()

	// Keyboard
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		i.prevKeyDurations[k] = i.keyDurations[k]
		if ebiten.IsKeyPressed(k) {
			i.keyDurations[k]++
		} else {
			i.keyDurations[k] = 0
		}
	}

	// Mouse
	for _, b := range []ebiten.MouseButton{
		ebiten.MouseButtonLeft,
		ebiten.MouseButtonRight,
		ebiten.MouseButtonMiddle,
	} {
		i.prevMouseButtonDurations[b] = i.mouseButtonDurations[b]
		if ebiten.IsMouseButtonPressed(b) {
			i.mouseButtonDurations[b]++
		} else {
			i.mouseButtonDurations[b] = 0
		}
	}

	// Gamepads

	// Copy the gamepad IDs
	i.prevGamepadIDs = map[int]struct{}{}
	for id := range i.gamepadIDs {
		i.prevGamepadIDs[id] = struct{}{}
	}

	// Reset the previous states first since some gamepad IDs might be already gone.
	for id := range i.prevGamepadButtonDurations {
		for b := range i.prevGamepadButtonDurations[id] {
			i.prevGamepadButtonDurations[id][b] = 0
		}
	}

	i.gamepadIDs = map[int]struct{}{}
	for _, id := range ebiten.GamepadIDs() {
		i.gamepadIDs[id] = struct{}{}

		if _, ok := i.prevGamepadButtonDurations[id]; !ok {
			i.prevGamepadButtonDurations[id] = map[ebiten.GamepadButton]int{}
		}
		if _, ok := i.gamepadButtonDurations[id]; !ok {
			i.gamepadButtonDurations[id] = map[ebiten.GamepadButton]int{}
		}

		n := ebiten.GamepadButtonNum(id)
		for b := ebiten.GamepadButton(0); b < ebiten.GamepadButton(n); b++ {
			i.prevGamepadButtonDurations[id][b] = i.gamepadButtonDurations[id][b]
			if ebiten.IsGamepadButtonPressed(id, b) {
				i.gamepadButtonDurations[id][b]++
			} else {
				i.gamepadButtonDurations[id][b] = 0
			}
		}
	}
	idsToDelete := []int{}
	for id := range i.gamepadButtonDurations {
		if _, ok := i.gamepadIDs[id]; !ok {
			idsToDelete = append(idsToDelete, id)
		}
	}
	for _, id := range idsToDelete {
		delete(i.gamepadButtonDurations, id)
	}

	// Touches
	ids := map[int]struct{}{}

	// Reset the previous states first since some gamepad IDs might be already gone.
	for id := range i.prevTouchDurations {
		i.prevTouchDurations[id] = 0
	}

	for _, t := range ebiten.Touches() {
		ids[t.ID()] = struct{}{}
		i.prevTouchDurations[t.ID()] = i.touchDurations[t.ID()]
		i.touchDurations[t.ID()]++
	}
	idsToDelete = []int{}
	for id := range i.touchDurations {
		if _, ok := ids[id]; !ok {
			idsToDelete = append(idsToDelete, id)
		}
	}
	for _, id := range idsToDelete {
		delete(i.touchDurations, id)
	}
}

// IsKeyJustPressed returns a boolean value indicating
// whether the given key is pressed just in the current frame.
//
// IsKeyJustPressed is concurrent safe.
func IsKeyJustPressed(key ebiten.Key) bool {
	return KeyPressDuration(key) == 1
}

// IsKeyJustReleased returns a boolean value indicating
// whether the given key is released just in the current frame.
//
// IsKeyJustReleased is concurrent safe.
func IsKeyJustReleased(key ebiten.Key) bool {
	theInputState.m.RLock()
	r := theInputState.keyDurations[key] == 0 && theInputState.prevKeyDurations[key] > 0
	theInputState.m.RUnlock()
	return r
}

// KeyPressDuration returns how long the key is pressed in frames.
//
// KeyPressDuration is concurrent safe.
func KeyPressDuration(key ebiten.Key) int {
	theInputState.m.RLock()
	s := theInputState.keyDurations[key]
	theInputState.m.RUnlock()
	return s
}

// IsMouseButtonJustPressed returns a boolean value indicating
// whether the given mouse button is pressed just in the current frame.
//
// IsMouseButtonJustPressed is concurrent safe.
func IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return MouseButtonPressDuration(button) == 1
}

// IsMouseButtonJustReleased returns a boolean value indicating
// whether the given mouse button is released just in the current frame.
//
// IsMouseButtonJustReleased is concurrent safe.
func IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	theInputState.m.RLock()
	r := theInputState.mouseButtonDurations[button] == 0 &&
		theInputState.prevMouseButtonDurations[button] > 0
	theInputState.m.RUnlock()
	return r
}

// MouseButtonPressDuration returns how long the mouse button is pressed in frames.
//
// MouseButtonPressDuration is concurrent safe.
func MouseButtonPressDuration(button ebiten.MouseButton) int {
	theInputState.m.RLock()
	s := theInputState.mouseButtonDurations[button]
	theInputState.m.RUnlock()
	return s
}

// JustConnectedGamepadIDs returns gamepad IDs that are connected just in the current frame.
//
// JustConnectedGamepadIDs might return nil when there is no connected gamepad.
//
// JustConnectedGamepadIDs is concurrent safe.
func JustConnectedGamepadIDs() []int {
	var ids []int
	theInputState.m.RLock()
	for id := range theInputState.gamepadIDs {
		if _, ok := theInputState.prevGamepadIDs[id]; !ok {
			ids = append(ids, id)
		}
	}
	theInputState.m.RUnlock()
	sort.Ints(ids)
	return ids
}

// JustDisconnectedGamepadIDs returns gamepad IDs that are disconnected just in the current frame.
//
// JustDisconnectedGamepadIDs might return nil when there is no disconnected gamepad.
//
// JustDisconnectedGamepadIDs is concurrent safe.
func JustDisconnectedGamepadIDs() []int {
	var ids []int
	theInputState.m.RLock()
	for id := range theInputState.prevGamepadIDs {
		if _, ok := theInputState.gamepadIDs[id]; !ok {
			ids = append(ids, id)
		}
	}
	theInputState.m.RUnlock()
	sort.Ints(ids)
	return ids
}

// IsGamepadButtonJustPressed returns a boolean value indicating
// whether the given gamepad button of the gamepad id is pressed just in the current frame.
//
// IsGamepadButtonJustPressed is concurrent safe.
func IsGamepadButtonJustPressed(id int, button ebiten.GamepadButton) bool {
	return GamepadButtonPressDuration(id, button) == 1
}

// IsGamepadButtonJustReleased returns a boolean value indicating
// whether the given gamepad button of the gamepad id is released just in the current frame.
//
// IsGamepadButtonJustReleased is concurrent safe.
func IsGamepadButtonJustReleased(id int, button ebiten.GamepadButton) bool {
	theInputState.m.RLock()
	prev := 0
	if _, ok := theInputState.prevGamepadButtonDurations[id]; ok {
		prev = theInputState.prevGamepadButtonDurations[id][button]
	}
	current := 0
	if _, ok := theInputState.gamepadButtonDurations[id]; ok {
		current = theInputState.gamepadButtonDurations[id][button]
	}
	theInputState.m.RUnlock()
	return current == 0 && prev > 0
}

// GamepadButtonPressDuration returns how long the gamepad button of the gamepad id is pressed in frames.
//
// GamepadButtonPressDuration is concurrent safe.
func GamepadButtonPressDuration(id int, button ebiten.GamepadButton) int {
	theInputState.m.RLock()
	s := 0
	if _, ok := theInputState.gamepadButtonDurations[id]; ok {
		s = theInputState.gamepadButtonDurations[id][button]
	}
	theInputState.m.RUnlock()
	return s
}

// JustPressedTouches returns touch IDs that are created just in the current frame.
//
// JustPressedTouches might return nil when there is not touch.
//
// JustPressedTouches is concurrent safe.
func JustPressedTouches() []int {
	var ids []int
	theInputState.m.RLock()
	for id, s := range theInputState.touchDurations {
		if s == 1 {
			ids = append(ids, id)
		}
	}
	theInputState.m.RUnlock()
	sort.Ints(ids)
	return ids
}

// IsTouchJustReleased returns a boolean value indicating
// whether the given touch is released just in the current frame.
//
// IsTouchJustReleased is concurrent safe.
func IsTouchJustReleased(id int) bool {
	theInputState.m.RLock()
	r := theInputState.touchDurations[id] == 0 && theInputState.prevTouchDurations[id] > 0
	theInputState.m.RUnlock()
	return r
}

// TouchDuration returns how long the touch remains in frames.
//
// TouchDuration is concurrent safe.
func TouchDuration(id int) int {
	theInputState.m.RLock()
	s := theInputState.touchDurations[id]
	theInputState.m.RUnlock()
	return s
}