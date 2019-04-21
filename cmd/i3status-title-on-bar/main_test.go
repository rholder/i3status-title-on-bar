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

/*

type Config struct {
	color      string
	appendEnd  bool
	fixedWidth int
	printHelp  bool
	printVersion bool
}

func TestCliNoArgs(t *testing.T) {
	args := []string{}
	in := strings.NewReader("")
	err := cli(args, in, ioutil.Discard, ioutil.Discard)
	if err == nil {
		t.Fatal("Expected an error")
	} else {
		if err.Error() != "Invalid number of arguments." {
			t.Fatal("Expected argument error")
		}
	}
}

func TestCliArgs(t *testing.T) {
	args := []string{"aaa"}
	in := strings.NewReader("potato\naaa\nmeep")
	err := cli(args, in, ioutil.Discard, ioutil.Discard)
	if err != nil {
		t.Fatal("Unexpected error")
	}

	// TODO add more tests to verify counts
}
*/
