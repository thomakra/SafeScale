/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"

	"github.com/sethvargo/go-password/password"
)

var generator *password.Generator

// GeneratePassword generates a password with length at least 12
func GeneratePassword(length uint8) (string, error) {
	if length < 12 {
		panic("length under 12!")
	}
	password, err := generator.Generate(int(length), 4, 4, false, true)
	if err != nil {
		return "", err
	}
	return password, nil
}

func init() {
	var err error
	// generator is created with characters allowed
	// potential confusing characters, like i/l/| or 0/O, are removed to ease human readability
	generator, err = password.NewGenerator(&password.GeneratorInput{
		LowerLetters: "abcdefghjkmnopqrstuvwxyz",
		UpperLetters: "ABCDEFGHJKLMNPQRSTUVWXYZ",
		Digits:       "123456789",
		Symbols:      "-+*/.,:()[]{}#_",
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create password generator: %s!", err.Error()))
	}
}
