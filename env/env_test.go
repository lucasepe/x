package env_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/lucasepe/x/env"
)

func ExampleStrs() {
	os.Setenv("ASDF", "aaa,bbb,ccc,ddd")
	defer env.Unset("ASDF")

	all := env.Strs("ASDF", ",")
	none := env.Strs("ASDF_1", ",")
	defs := env.Strs("ASDF_1", ",", "one", "two", "three!")

	fmt.Println(all)
	fmt.Println(none)
	fmt.Println(defs)

	// Output:
	// [aaa bbb ccc ddd]
	// []
	// [one two three!]
}

func TestContains(t *testing.T) {
	env.Unset("ASDF")
	os.Setenv("ASDF", "123asdf123")
	if !env.Contains("ASDF", "asdf") {
		t.Fail()
	}
	env.Unset("ASDF")
}

func TestEqual(t *testing.T) {
	env.Unset("ASDFASDFASDF")
	if env.Str("ASDFASDFASDF", "") != "" {
		t.Fail()
	}
	os.Setenv("ASDFASDFASDF", "1")
	if !env.Equal("ASDFASDFASDF", "1") {
		t.Fail()
	}
	env.Unset("ASDFASDFASDF")
}

func TestBool(t *testing.T) {
	env.Unset("QWERTYQWERTY")
	if env.True("QWERTYQWERTY") {
		t.Fail()
	}
	os.Setenv("QWERTYQWERTY", "yes")
	if !env.True("QWERTYQWERTY") {
		t.Fail()
	}
	env.Unset("QWERTYQWERTY")
}

func TestExpandUser(t *testing.T) {
	if env.ExpandUser("~/test") == "/tmp" {
		t.Fail()
	}
}

func TestPath(t *testing.T) {
	os.Setenv("PATH", "/usr/bin:/bin:/usr/local/bin")
	if len(env.Strs("PATH", string(os.PathListSeparator))) != 3 {
		t.Fail()
	}
	env.Unset("PATH")
}

func TestHomeDir(t *testing.T) {
	if env.HomeDir() == "" {
		t.Fail()
	}
}

func TestFloat32(t *testing.T) {
	os.Setenv("F", "1.234")
	f := env.Float32("F", 0.0)
	if f != 1.234 {
		t.Fail()
	}
	env.Unset("F")
}
