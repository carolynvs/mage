package sh

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/magefile/mage/mg"
)

func TestPreparedCommand_Run(t *testing.T) {
	stdout := CaptureStdout()
	defer stdout.Release()

	c := Command("go", "run", "echo.go", "hello world")
	err := c.Run()
	if err != nil {
		t.Fatal(err)
	}

	got := stdout.Output()
	want := ""
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_Run_Verbose(t *testing.T) {
	os.Setenv(mg.VerboseEnv, "true")
	defer os.Unsetenv(mg.VerboseEnv)

	stdout := CaptureStdout()
	defer stdout.Release()

	c := Command("go", "run", "echo.go", "hello world")
	err := c.Run()
	if err != nil {
		t.Fatal(err)
	}

	got := stdout.Output()
	want := "hello world\n"
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_Run_Fail(t *testing.T) {
	stderr := CaptureStderr()
	defer stderr.Release()

	c := Command("go", "run")
	err := c.Run()
	if err == nil {
		t.Fatalf("expected %s to fail", c)
	}

	got := stderr.Output()

	want := "go run: no go files listed\n"
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_RunS(t *testing.T) {
	buf := &bytes.Buffer{}
	c := Command("go", "run", "echo.go", "hello world")
	c.Cmd.Stdout = buf

	err := c.RunS()
	if err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	want := ""
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_RunS_Fail(t *testing.T) {
	stderr := CaptureStderr()
	defer stderr.Release()

	c := Command("go", "run")
	err := c.RunS()
	if err == nil {
		t.Fatalf("expected %s to fail", c)
	}

	got := stderr.Output()

	want := ""
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_RunE(t *testing.T) {
	stderr := CaptureStderr()
	defer stderr.Release()

	c := Command("go", "run", "echo.go", "hello world")
	err := c.RunE()
	if err != nil {
		t.Fatal(err)
	}

	got := stderr.Output()

	want := ""
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_RunE_Fail(t *testing.T) {
	stderr := CaptureStderr()
	defer stderr.Release()

	c := Command("go", "run")
	err := c.RunE()
	if err == nil {
		t.Fatalf("expected %s to fail", c)
	}

	got := stderr.Output()

	want := "go run: no go files listed\n"
	if got != want {
		t.Fatalf("want: %q got: %q", want, got)
	}
}

func TestPreparedCommand_Output(t *testing.T) {
	/*
		stdout := CaptureStdout()
		defer stdout.Release()

			gotOutput, err := Command("go", "run", "echo.go", "hello world").Output()
			if err != nil {
				t.Fatal(err)
			}

			wantOutput := "hello world"
			if gotOutput != wantOutput {
				t.Fatalf("wantOutput: %q gotOutput: %q", wantOutput, gotOutput)
			}

				gotStdout := stdout.Output()
				wantStdout := ""
				if gotStdout != wantStdout {
					t.Fatalf("wantStdout: %q gotStdout: %q", wantStdout, gotStdout)
				}
	*/
}

func TestPreparedCommand_Output_Verbose(t *testing.T) {
	os.Setenv(mg.VerboseEnv, "true")
	defer os.Unsetenv(mg.VerboseEnv)

	stdout := CaptureStdout()
	defer stdout.Release()

	gotOutput, err := Command("go", "run", "echo.go", "hello world").Output()
	if err != nil {
		t.Fatal(err)
	}

	wantOutput := "hello world"
	if gotOutput != wantOutput {
		t.Fatalf("wantOutput: %q gotOutput: %q", wantOutput, gotOutput)
	}

	gotStdout := stdout.Output()
	wantStdout := "hello world\n"
	if gotStdout != wantStdout {
		t.Fatalf("wantStdout: %q gotStdout: %q", wantStdout, gotStdout)
	}
}

func ExamplePreparedCommand_RunV() {
	err := Command("go", "run", "echo.go", "hello world").RunV()
	if err != nil {
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
	err = Command("go", "run", "test_main.go").Stdout(os.Stdout).In(tmp).RunV()
	if err != nil {
		log.Fatal(err)
	}
	// Output: hello world
}

func ExamplePreparedCommand_Silent() {
	err := Command("go", "run", "echo.go", "hello world").Silent().Run()
	if err != nil {
		log.Fatal(err)
	}
	// Output:
}
