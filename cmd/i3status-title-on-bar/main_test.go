// Copyright 2019 Ray Holder
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

package main

import (
	"io/ioutil"
	"testing"
)

func TestCliNoArgs(t *testing.T) {
	args := []string{}
	config, err := newConfig("test", args)
	if err != nil {
		t.Fatal("Unexpected error")
	}
	if config.color != "#00FF00" {
		t.Fatal("Expected default color")
	}
	if config.appendEnd {
		t.Fatal("Unexpected appendEnd default")
	}
	if config.fixedWidth != 0 {
		t.Fatal("Unexpected fixedWidth default")
	}
	if config.printHelp {
		t.Fatal("Unexpected printHelp default")
	}
	if config.printVersion {
		t.Fatal("Unexpected printVersion default")
	}

	exit, code := shouldExit(ioutil.Discard, config, err)
	if exit {
		t.Fatal("Unexpected exit")
	}
	if code != 0 {
		t.Fatal("Unexpected exit code")
	}
}

func TestCliVersionArgs(t *testing.T) {
	args := []string{"--version"}
	config, err := newConfig("test", args)
	if err != nil {
		t.Fatal("Unexpected error")
	}
	if config.color != "#00FF00" {
		t.Fatal("Expected default color")
	}
	if config.appendEnd {
		t.Fatal("Unexpected appendEnd default")
	}
	if config.fixedWidth != 0 {
		t.Fatal("Unexpected fixedWidth default")
	}
	if config.printHelp {
		t.Fatal("Unexpected printHelp default")
	}
	if !config.printVersion {
		t.Fatal("Unexpected printVersion")
	}

	exit, code := shouldExit(ioutil.Discard, config, err)
	if !exit {
		t.Fatal("Unexpected exit")
	}
	if code == 0 {
		t.Fatal("Unexpected exit code")
	}
}

func TestCliHelpArgs(t *testing.T) {
	args := []string{"--help"}
	config, err := newConfig("test", args)
	if err != nil {
		t.Fatal("Unexpected error")
	}
	if config.color != "#00FF00" {
		t.Fatal("Expected default color")
	}
	if config.appendEnd {
		t.Fatal("Unexpected appendEnd default")
	}
	if config.fixedWidth != 0 {
		t.Fatal("Unexpected fixedWidth default")
	}
	if !config.printHelp {
		t.Fatal("Unexpected printHelp")
	}
	if config.printVersion {
		t.Fatal("Unexpected printVersion default")
	}

	exit, code := shouldExit(ioutil.Discard, config, err)
	if !exit {
		t.Fatal("Unexpected exit")
	}
	if code == 0 {
		t.Fatal("Unexpected exit code")
	}
}

func TestCliBadArgs(t *testing.T) {
	args := []string{"--bad-args"}
	config, err := newConfig("test", args)
	if err == nil {
		t.Fatal("Expected error")
	}
	if config.color != "#00FF00" {
		t.Fatal("Expected default color")
	}
	if config.appendEnd {
		t.Fatal("Unexpected appendEnd default")
	}
	if config.fixedWidth != 0 {
		t.Fatal("Unexpected fixedWidth default")
	}
	if config.printHelp {
		t.Fatal("Unexpected printHelp default")
	}
	if config.printVersion {
		t.Fatal("Unexpected printVersion default")
	}

	exit, code := shouldExit(ioutil.Discard, config, err)
	if !exit {
		t.Fatal("Unexpected exit")
	}
	if code == 0 {
		t.Fatal("Unexpected exit code")
	}
}
