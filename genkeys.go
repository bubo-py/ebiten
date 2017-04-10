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
	"log"
	"os"
	"sort"
	"strconv"
	"text/template"

	"github.com/hajimehoshi/ebiten/internal"
)

var keyCodeToName map[string]string

func init() {
	keyCodeToName = map[string]string{
		"Comma":        "Comma",
		"Period":       "Period",
		"AltLeft":      "Alt",
		"AltRight":     "Alt",
		"CapsLock":     "CapsLock",
		"ControlLeft":  "Control",
		"ControlRight": "Control",
		"ShiftLeft":    "Shift",
		"ShiftRight":   "Shift",
		"Enter":        "Enter",
		"Space":        "Space",
		"Tab":          "Tab",
		"Delete":       "Delete",
		"End":          "End",
		"Home":         "Home",
		"Insert":       "Insert",
		"PageDown":     "PageDown",
		"PageUp":       "PageUp",
		"ArrowDown":    "Down",
		"ArrowLeft":    "Left",
		"ArrowRight":   "Right",
		"ArrowUp":      "Up",
		"Escape":       "Escape",
		"Backspace":    "Backspace",
	}
	// ASCII: 0 - 9
	for c := '0'; c <= '9'; c++ {
		keyCodeToName["Digit"+string(c)] = string(c)
	}
	// ASCII: A - Z
	for c := 'A'; c <= 'Z'; c++ {
		keyCodeToName["Key"+string(c)] = string(c)
	}
	// Function keys
	for i := 1; i <= 12; i++ {
		keyCodeToName["F"+strconv.Itoa(i)] = "F" + strconv.Itoa(i)
	}
}

const ebitenKeysTmpl = `{{.License}}

// {{.Notice}}

package ebiten

import (
	"github.com/hajimehoshi/ebiten/internal/ui"
)

// A Key represents a keyboard key.
// These keys represent pysical keys of US keyboard.
// For example, KeyQ represents Q key on US keybaords and ' (quote) key on Dvorak keyboards.
type Key int

// Keys
const (
{{range $index, $name := .KeyNames}}Key{{$name}} = Key(ui.Key{{$name}})
{{end}}	KeyMax = Key{{.LastKeyName}}
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

{{.BuildTag}}

package ui

import (
	glfw "github.com/go-gl/glfw/v3.2/glfw"
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

{{.BuildTag}}

package ui

var keyCodeToKey = map[string]Key{
{{range $code, $name := .KeyCodeToName}}"{{$code}}": Key{{$name}},
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

	namesSet := map[string]struct{}{}
	namesWithoutModsSet := map[string]struct{}{}
	codes := []string{}
	for code, name := range keyCodeToName {
		namesSet[name] = struct{}{}
		codes = append(codes, code)
		if name != "Alt" && name != "Control" && name != "Shift" {
			namesWithoutModsSet[name] = struct{}{}
		}
	}
	names := []string{}
	namesWithoutMods := []string{}
	for n := range namesSet {
		names = append(names, n)
	}
	for n := range namesWithoutModsSet {
		namesWithoutMods = append(namesWithoutMods, n)
	}

	sort.Sort(KeyNames(names))
	sort.Sort(KeyNames(namesWithoutMods))
	sort.Strings(codes)

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
		// The build tag can't be included in the templates because of `go vet`.
		// Pass the build tag and extract this in the template to make `go vet` happy.
		buildTag := ""
		switch path {
		case "internal/ui/keys_glfw.go":
			buildTag = "// +build darwin linux windows" +
				"\n// +build !js" +
				"\n// +build !android" +
				"\n// +build !ios"
		case "internal/ui/keys_js.go":
			buildTag = "// +build js"
		}
		// NOTE: According to godoc, maps are automatically sorted by key.
		if err := tmpl.Execute(f, map[string]interface{}{
			"License":             license,
			"Notice":              notice,
			"BuildTag":            buildTag,
			"KeyCodeToName":       keyCodeToName,
			"Codes":               codes,
			"KeyNames":            names,
			"LastKeyName":         names[len(names)-1],
			"KeyNamesWithoutMods": namesWithoutMods,
		}); err != nil {
			log.Fatal(err)
		}
	}
}
