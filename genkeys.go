// Copyright 2015 Hajime Hoshi
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

// +build ignore

// Note:
//   * Respect GLFW key names
//   * https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent.keyCode
//   * It is best to replace keyCode with code, but many browsers don't implement it.

package main

import (
	"github.com/hajimehoshi/ebiten/internal"
	"log"
	"os"
	"sort"
	"strconv"
	"text/template"
)

var keyCodeToName map[int]string

func init() {
	keyCodeToName = map[int]string{
		0xBC: "Comma",
		0xBE: "Period",
		0x12: "Alt",
		0x14: "CapsLock",
		0x11: "Control",
		0x10: "Shift",
		0x0D: "Enter",
		0x20: "Space",
		0x09: "Tab",
		0x2E: "Delete",
		0x23: "End",
		0x24: "Home",
		0x2D: "Insert",
		0x22: "PageDown",
		0x21: "PageUp",
		0x28: "Down",
		0x25: "Left",
		0x27: "Right",
		0x26: "Up",
		0x1B: "Escape",
		// The keys not listed in the Mozilla site:
		0x08: "Backspace",
	}
	// ASCII: 0 - 9
	for c := '0'; c <= '9'; c++ {
		keyCodeToName[int(c)] = string(c)
	}
	// ASCII: A - Z
	for c := 'A'; c <= 'Z'; c++ {
		keyCodeToName[int(c)] = string(c)
	}
	// Function keys
	for i := 1; i <= 12; i++ {
		keyCodeToName[0x70+i-1] = "F" + strconv.Itoa(i)
	}
}

const ebitenKeysTmpl = `{{.License}}

// {{.Notice}}

package ebiten


import (
	"github.com/hajimehoshi/ebiten/internal/ui"
)

// A Key represents a keyboard key.
type Key int

// Keys
const (
{{range $index, $name := .KeyNames}}Key{{$name}} = Key(ui.Key{{$name}})
{{end}}
)
`

const uiKeysTmpl = `{{.License}}

// {{.Notice}}

package ui

type Key int

const (
{{range $index, $name := .KeyNames}}Key{{$name}}{{if eq $index 0}} Key = iota{{end}}
{{end}}
)
`

const uiKeysGlfwTmpl = `{{.License}}

// {{.Notice}}

// +build !js

package ui

import (
	glfw "github.com/go-gl/glfw3"
)

var glfwKeyCodeToKey = map[glfw.Key]Key{
{{range $index, $name := .KeyNamesWithoutMods}}glfw.Key{{$name}}: Key{{$name}},
{{end}}
	glfw.KeyLeftAlt:      KeyAlt,
	glfw.KeyRightAlt:     KeyAlt,
	glfw.KeyLeftControl:  KeyControl,
	glfw.KeyRightControl: KeyControl,
	glfw.KeyLeftShift:    KeyShift,
	glfw.KeyRightShift:   KeyShift,
}
`

const uiKeysJSTmpl = `{{.License}}

// {{.Notice}}

// +build js

package ui

var keyCodeToKey = map[int]Key{
{{range $code, $name := .KeyCodeToName}}{{$code}}: Key{{$name}},
{{end}}
}
`

type KeyNames []string

func (k KeyNames) digit(name string) int {
	if len(name) != 1 {
		return -1
	}
	c := name[0]
	if c < '0' || '9' < c {
		return -1
	}
	return int(c - '0')
}

func (k KeyNames) alphabet(name string) rune {
	if len(name) != 1 {
		return -1
	}
	c := rune(name[0])
	if c < 'A' || 'Z' < c {
		return -1
	}
	return c
}

func (k KeyNames) function(name string) int {
	if len(name) < 2 {
		return -1
	}
	if name[0] != 'F' {
		return -1
	}
	i, err := strconv.Atoi(name[1:])
	if err != nil {
		return -1
	}
	return i
}

func (k KeyNames) Len() int {
	return len(k)
}

func (k KeyNames) Less(i, j int) bool {
	k0, k1 := k[i], k[j]
	d0, d1 := k.digit(k0), k.digit(k1)
	a0, a1 := k.alphabet(k0), k.alphabet(k1)
	f0, f1 := k.function(k0), k.function(k1)
	if d0 != -1 {
		if d1 != -1 {
			return d0 < d1
		}
		return true
	}
	if a0 != -1 {
		if d1 != -1 {
			return false
		}
		if a1 != -1 {
			return a0 < a1
		}
		return true
	}
	if d1 != -1 {
		return false
	}
	if a1 != -1 {
		return false
	}
	if f0 != -1 && f1 != -1 {
		return f0 < f1
	}
	return k0 < k1
}

func (k KeyNames) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func main() {
	license, err := internal.LicenseComment()
	if err != nil {
		log.Fatal(err)
	}

	notice := "DO NOT EDIT: This file is auto-generated by genkeys.go."

	names := []string{}
	namesWithoutMods := []string{}
	codes := []int{}
	for code, name := range keyCodeToName {
		names = append(names, name)
		codes = append(codes, code)
		if name != "Alt" && name != "Control" && name != "Shift" {
			namesWithoutMods = append(namesWithoutMods, name)
		}
	}
	sort.Sort(KeyNames(names))
	sort.Sort(KeyNames(namesWithoutMods))
	sort.Ints(codes)

	for path, tmpl := range map[string]string{
		"keys.go":                  ebitenKeysTmpl,
		"internal/ui/keys.go":      uiKeysTmpl,
		"internal/ui/keys_glfw.go": uiKeysGlfwTmpl,
		"internal/ui/keys_js.go":   uiKeysJSTmpl,
	} {
		f, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		tmpl, err := template.New(path).Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}
		// NOTE: According to godoc, maps are automatically sorted by key.
		if err := tmpl.Execute(f, map[string]interface{}{
			"License":             license,
			"Notice":              notice,
			"KeyCodeToName":       keyCodeToName,
			"Codes":               codes,
			"KeyNames":            names,
			"KeyNamesWithoutMods": namesWithoutMods,
		}); err != nil {
			log.Fatal(err)
		}
	}
}
