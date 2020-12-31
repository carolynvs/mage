package sh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
)

// RunCmd returns a function that will call Run with the given command. This is
// useful for creating command aliases to make your scripts easier to read, like
// this:
//
//  // in a helper file somewhere
//  var g0 = sh.RunCmd("go")  // go is a keyword :(
//
//  // somewhere in your main code
//	if err := g0("install", "github.com/gohugo/hugo"); err != nil {
//		return err
//  }
//
// Args passed to command get baked in as args to the command when you run it.
// Any args passed in when you run the returned function will be appended to the
// original args.  For example, this is equivalent to the above:
//
//  var goInstall = sh.RunCmd("go", "install") goInstall("github.com/gohugo/hugo")
//
// RunCmd uses Exec underneath, so see those docs for more details.
func RunCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return Run(cmd, append(args, args2...)...)
	}
}

// OutCmd is like RunCmd except the command returns the output of the
// command.
func OutCmd(cmd string, args ...string) func(args ...string) (string, error) {
	return func(args2 ...string) (string, error) {
		return Output(cmd, append(args, args2...)...)
	}
}

// Run is like RunWith, but doesn't specify any environment variables.
func Run(cmd string, args ...string) error {
	return RunWith(nil, cmd, args...)
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func RunV(cmd string, args ...string) error {
	_, err := Exec(nil, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// RunWith runs the given command, directing stderr to this program's stderr and
// printing stdout to stdout if mage was run with -v.  It adds adds env to the
// environment variables for the command being run. Environment variables should
// be in the format name=value.
func RunWith(env map[string]string, cmd string, args ...string) error {
	var output io.Writer
	if mg.Verbose() {
		output = os.Stdout
	}
	_, err := Exec(env, output, os.Stderr, cmd, args...)
	return err
}

// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	_, err := Exec(env, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// Output runs the command and returns the text from stdout.
func Output(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(nil, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// OutputWith is like RunWith, but returns what is written to stdout.
func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(env, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// Exec executes the command, piping its stderr to mage's stderr and
// piping its stdout to the given writer. If the command fails, it will return
// an error that, if returned from a target or mg.Deps call, will cause mage to
// exit with the same code as the command failed with.  Env is a list of
// environment variables to set when running the command, these override the
// current environment variables set (which are also passed to the command). cmd
// and args may include references to environment variables in $FOO format, in
// which case these will be expanded before the command is run.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	expand := func(s string) string {
		s2, ok := env[s]
		if ok {
			return s2
		}
		return os.Getenv(s)
	}
	cmd = os.Expand(cmd, expand)
	for i := range args {
		args[i] = os.Expand(args[i], expand)
	}
	ran, code, err := run(env, stdout, stderr, cmd, args...)
	if err == nil {
		return true, nil
	}
	if ran {
		return ran, mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
	}
	return ran, fmt.Errorf(`failed to run "%s %s: %v"`, cmd, strings.Join(args, " "), err)
}

func run(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, code int, err error) {
	c := PreparedCommand{exec.Command(cmd, args...)}
	c.Cmd.Env = os.Environ()
	for k, v := range env {
		c.Cmd.Env = append(c.Cmd.Env, k+"="+v)
	}
	c.Cmd.Stderr = stderr
	c.Cmd.Stdout = stdout
	c.Cmd.Stdin = os.Stdin
	return c.Run()
}

type PreparedCommand struct {
	Cmd *exec.Cmd
}

// Command creates a default command. Stdout is logged in verbose mode. Stderr
// is sent to os.Stderr.
func Command(cmd string, args ...string) PreparedCommand {
	var output io.Writer
	if mg.Verbose() {
		output = os.Stdout
	}
	c := exec.Command(cmd, args...)
	c.Stdout = output
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	return PreparedCommand{Cmd: c}
}

// Args appends additional arguments to the command.
func (c PreparedCommand) Args(args ...string) PreparedCommand {
	c.Cmd.Args = append(c.Cmd.Args, args...)
	return c
}

// Env defines additional environment variables for the command.
// All ambient environment variables are included by default.
// Example:
//  c.Env("X=1", "Y=2")
func (c PreparedCommand) Env(vars ...string) PreparedCommand {
	for _, v := range vars {
		c.Cmd.Env = append(c.Cmd.Env, v)
	}
	return c
}

// In sets the working directory of the command.
func (c PreparedCommand) In(dir string) PreparedCommand {
	c.Cmd.Dir = dir
	return c
}

// Stdout directs stdout from the command.
func (c PreparedCommand) Stdout(stdout io.Writer) PreparedCommand {
	c.Cmd.Stdout = stdout
	return c
}

// Stderr directs stderr from the command.
func (c PreparedCommand) Stderr(stdout io.Writer) PreparedCommand {
	c.Cmd.Stdout = stdout
	return c
}

// Runs a command silently, without logging stdout/stderr.
func (c PreparedCommand) Silent() PreparedCommand {
	c.Cmd.Stdout = nil
	c.Cmd.Stderr = nil
	return c
}

// Run executes the prepared command, returning if the command was run and its
// exit code.
func (c PreparedCommand) Run() (ran bool, code int, err error) {
	if mg.Verbose() {
		log.Println("exec:", c.Cmd.Path, strings.Join(c.Cmd.Args, " "))
	}
	err = c.Cmd.Run()

	env := make(map[string]string, len(c.Cmd.Env))
	for _, envVar := range c.Cmd.Env {
		parts := strings.SplitAfterN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		env[parts[0]] = parts[1]
	}
	// TODO: implement sh.Exec here
	expand := func(s string) string {
		return env[s]
	}
	cmd = os.Expand(cmd, expand)
	for i := range args {
		args[i] = os.Expand(args[i], expand)
	}
	ran, code, err := run(env, stdout, stderr, cmd, args...)
	if err == nil {
		return true, nil
	}
	if ran {
		return ran, mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
	}
	return ran, fmt.Errorf(`failed to run "%s %s: %v"`, cmd, strings.Join(args, " "), err)

	return CmdRan(err), ExitStatus(err), err
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func (c PreparedCommand) RunV() error {
	c.Stdout(os.Stdout)
	_, _, err := c.Run()
	return err
}

// Output executes the prepared command, returning stdout.
func (c PreparedCommand) Output() (string, error) {
	buf := &bytes.Buffer{}
	c.Cmd.Stdout = buf
	_, _, err := c.Run()
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// CmdRan examines the error to determine if it was generated as a result of a
// command running via os/exec.Command.  If the error is nil, or the command ran
// (even if it exited with a non-zero exit code), CmdRan reports true.  If the
// error is an unrecognized type, or it is an error from exec.Command that says
// the command failed to run (usually due to the command not existing or not
// being executable), it reports false.
func CmdRan(err error) bool {
	if err == nil {
		return true
	}
	ee, ok := err.(*exec.ExitError)
	if ok {
		return ee.Exited()
	}
	return false
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or 1 if it is a different error.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(exitStatus); ok {
		return e.ExitStatus()
	}
	if e, ok := err.(*exec.ExitError); ok {
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}
	return 1
}
