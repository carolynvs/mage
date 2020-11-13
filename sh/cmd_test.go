package sh

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestOutCmd(t *testing.T) {
	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd("baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Errorf("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, os.Args[0], "-printVar", env)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("MAGE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}

}

func TestPreparedCommand_Run(t *testing.T) {
	buf := &bytes.Buffer{}
	c := Command("bash", "-c", "echo hello world")
	c.Cmd.Stdout = buf

	_, _, err := c.Run()
	if err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	want := "hello world\n"
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_Output(t *testing.T) {
	got, err := Command("bash", "-c", "echo hello world").Output()
	if err != nil {
		t.Fatal(err)
	}

	want := "hello world"
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func ExamplePreparedCommand_Stdout() {
	_, _, err := Command("bash", "-c", "echo hello world").Stdout(os.Stdout).Run()
	if err != nil{
		log.Fatal(err)
	}
	// Output: hello world
}

func ExamplePreparedCommand_In() {
	tmp, err := ioutil.TempDir("", "mage")
	if err != nil {
		log.Fatal(err)
	}

	contents := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`
	err = ioutil.WriteFile(filepath.Join(tmp, "test_main.go"), []byte(contents), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Run `go run test_main.go` in /tmp
	_, _, err = Command("go", "run", "test_main.go").Stdout(os.Stdout).In(tmp).Run()
	if err != nil {
		log.Fatal(err)
	}
	// Output: hello world
}

func ExamplePreparedCommand_Silent() {
	_, _, err := Command("bash", "-c", "echo hello world").Silent().Run()
	if err != nil {
		log.Fatal(err)
	}
	// Output:
}
