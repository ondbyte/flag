package flag_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/ondbyte/turbo_flag"
)

func Test_GetFirstSubCommandWithArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
		want2 bool
	}{
		{
			name: "sub command exists",
			args: args{
				args: []string{"yadu", "turbo", "--yes"},
			},
			want:  "yadu",
			want1: []string{"turbo", "--yes"},
			want2: true,
		},
		{
			name: "sub command doesnt exist",
			args: args{
				args: []string{"--yadu", "--turbo", "--yes"},
			},
			want:  "",
			want1: nil,
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := GetFirstSubCommandWithArgs(tt.args.args)
			if got != tt.want {
				t.Errorf("GetFirstSubCommandWithArgs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetFirstSubCommandWithArgs() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GetFirstSubCommandWithArgs() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestFlagSet_BindEnv(t *testing.T) {
	fs := OneCmd("test", ContinueOnError)
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	f, err := os.Create(envFile)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = f.WriteString(`
	YADU=123
	PORT=3555
	YES=TRUE
	`)
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = fs.LoadEnv(envFile)
	if err != nil {
		t.Fatalf("failed to load env : %v", err)
	}
	name := fs.String("str", "", "", fs.Env("YADU"))
	port := fs.Int("port", 0, "", fs.Env("PORT"))
	yes := fs.Bool("yes", false, "", fs.Env("YES"))
	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("ecpected error to be nil %v", err)
		}
	}()

	if *name != "123" {
		t.Fatal("name should have the value")
	}
	if *port != 3555 {
		t.Fatal("port should have the value")
	}
	if !*yes {
		t.Fatal("yes should have the value")
	}
}

func TestFlagSet_BindCfg(t *testing.T) {
	for _, ext := range []string{"json", "yaml", "yml", "toml"} {
		defer func() {
			err := recover()
			if err != nil {
				t.Fatalf("expected error to be nil %v", err)
			}
		}()
		path := "./test_config/demo." + ext
		fs := OneCmd("test", ContinueOnError)
		err := fs.LoadCfg(path)
		if err != nil {
			t.Fatalf("cfg [%v] file should exist", path)
		}
		password := fs.String("password", "", "", fs.Cfg("database.password"))

		if *password != "12345" {
			t.Fatal("expected password to be 12345")
		}
	}
	for _, ext := range []string{"json22", "yam2l", "ym2l", "tom2l"} {
		fs := OneCmd("test", ContinueOnError)
		password := fs.String("password", "", "", fs.Cfg("database.password"))
		defer func() {
			err := recover()
			if err != nil {
				t.Fatalf("expected error to be  nil %v", err)
			}
		}()

		err := fs.LoadCfg("./test_config/demo." + ext)
		if err == nil {
			t.Fatal("expected err")
		}

		if *password != "" {
			t.Fatalf("expected password to be empty but %v", *password)
		}

		if *password != "" {
			t.Fatal("expected password to be empty")
		}
	}
}

func TestFlagSet_Alias(t *testing.T) {
	fs := OneCmd("test", ContinueOnError)
	password := fs.String("password", "", "", fs.Alias("p"))
	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("ecpected error to be nil %v", err)
		}
	}()

	err := fs.Parse([]string{"-p", "12345"})
	if err != nil {
		t.Fatal("expected no error")
	}
	if *password != "12345" {
		t.Fatal("expected password to be 12345")
	}
}

func TestFlagSet_Enums(t *testing.T) {
	fs := OneCmd("test", ContinueOnError)
	password := fs.String("password", "12345", "", fs.Enum("12345", "123456789"))
	defer func() {
		err := recover()
		if err != nil {
			t.Fatalf("ecpected error to be nil %v", err)
		}
	}()

	options := fs.String("options", "c", "", fs.Enum("a", "b", "c"), fs.Alias("o"), fs.Env("OPTIONS", "OPTIONS2"))

	err := fs.Parse([]string{"-options", "z"})
	if err == nil {
		t.Fatal("expected error")
	}
	err = fs.Parse([]string{"-options", "a"})
	if err != nil {
		t.Fatal("expected no error")
	}
	if *password != "12345" {
		t.Fatal("expected password to be 12345")
	}
	if *options != "a" {
		t.Fatal("expected option to be a")
	}
}

func boolString(s string) string {
	if s == "0" {
		return "false"
	}
	return "true"
}

func testParse(f *Command, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")
	extra := "one-extra-argument"
	args := []string{
		"-bool",
		"-bool2=true",
		"--int", "22",
		"--int64", "0x23",
		"-uint", "24",
		"--uint64", "25",
		"-string", "hello",
		"-float64", "2718e28",
		"-duration", "2m",
		extra,
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
	if len(f.Args()) != 1 {
		t.Error("expected one argument, got", len(f.Args()))
	} else if f.Args()[0] != extra {
		t.Errorf("expected argument %q got %q", extra, f.Args()[0])
	}
}

// Declare a user-defined flag type.
type flagVar []string

func (f *flagVar) String() string {
	return fmt.Sprint([]string(*f))
}

func (f *flagVar) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func TestUserDefined(t *testing.T) {
	var flags Command
	flags.Init("test", ContinueOnError)
	var v flagVar
	flags.Var(&v, "v", "usage")
	if err := flags.Parse([]string{"-v", "1", "-v", "2", "-v=3"}); err != nil {
		t.Error(err)
	}
	if len(v) != 3 {
		t.Fatal("expected 3 args; got ", len(v))
	}
	expect := "[1 2 3]"
	if v.String() != expect {
		t.Errorf("expected value %q got %q", expect, v.String())
	}
}

// Declare a user-defined boolean flag type.
type boolFlagVar struct {
	count int
}

func (b *boolFlagVar) String() string {
	return fmt.Sprintf("%d", b.count)
}

func (b *boolFlagVar) Set(value string) error {
	if value == "true" {
		b.count++
	}
	return nil
}

func (b *boolFlagVar) IsBoolFlag() bool {
	return b.count < 4
}

func TestUserDefinedBool(t *testing.T) {
	var flags Command
	flags.Init("test", ContinueOnError)
	var b boolFlagVar
	var err error
	flags.Var(&b, "b", "usage")
	if err = flags.Parse([]string{"-b", "-b", "-b", "-b=true", "-b=false", "-b", "barg", "-b"}); err != nil {
		if b.count < 4 {
			t.Error(err)
		}
	}

	if b.count != 4 {
		t.Errorf("want: %d; got: %d", 4, b.count)
	}

	if err == nil {
		t.Error("expected error; got none")
	}
}

func TestSetOutput(t *testing.T) {
	var flags Command
	var buf bytes.Buffer
	flags.SetOutput(&buf)
	flags.Init("test", ContinueOnError)
	flags.Parse([]string{"-unknown"})
	if out := buf.String(); !strings.Contains(out, "-unknown") {
		t.Logf("expected output mentioning unknown; got %q", out)
	}
}

// Test that -help invokes the usage message and returns ErrHelp.
func TestHelp(t *testing.T) {
	var helpCalled = false
	fs := OneCmd("help test", ContinueOnError)
	var flag bool
	fs.BoolVar(&flag, "flag", false, "regular flag")
	// Regular flag invocation should work
	err := fs.Parse([]string{"-flag=true"})
	if err != nil {
		t.Fatal("expected no err")
	}
	if !flag {
		t.Error("flag was not set by -flag")
	}
	// Help flag should work as expected.
	err = fs.Parse([]string{"-help"})
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != fmt.Errorf("flag provided but not defined: -help").Error() {
		t.Fatal("expected ErrHelp; got ", err)
	}
	// If we define a help flag, that should override.
	var help bool
	fs.BoolVar(&help, "help", false, "help flag")
	helpCalled = false
	err = fs.Parse([]string{"-help"})
	if err != nil {
		t.Fatal("expected no error for defined -help; got ", err)
	}
	if helpCalled {
		t.Fatal("help was called; should not have been for defined help flag")
	}
}

const defaultOutput = `"  -A\tfor bootstrapping, allow 'any' type\thas no default value\n  -Alongflagname\ndisable bounds checking\thas no default value\n  -C\ta boolean defaulting to true\tdefaults to [true]\n  -D path\nset relative path for local imports\thas no default value\n  -E string\nissue 23543\tdefaults to [0]\n  -F number\na non-zero number\tdefaults to [2.7]\n  -G float\na float that defaults to zero\thas no default value\n  -M string\na multiline\n    \thelp\n    \tstring\thas no default value\n  -N int\na non-zero int\tdefaults to [27]\n  -O\ta flag\n    \tmultiline help string\tdefaults to [true]\n  -Z int\nan int that defaults to zero\thas no default value\n  -maxT timeout\nset timeout for dial\thas no default value\n"`

func mustPanic(t *testing.T, testName string, expected string, f func()) {
	t.Helper()
	defer func() {
		switch msg := recover().(type) {
		case nil:
			t.Errorf("%s\n: expected panic(%q), but did not panic", testName, expected)
		case string:
			if msg != expected {
				t.Errorf("%s\n: expected panic(%q), but got panic(%q)", testName, expected, msg)
			}
		default:
			t.Errorf("%s\n: expected panic(%q), but got panic(%T%v)", testName, expected, msg, msg)
		}
	}()
	f()
}

func TestInvalidFlags(t *testing.T) {
	tests := []struct {
		flag     string
		errorMsg string
	}{
		{
			flag:     "-foo",
			errorMsg: "flag \"-foo\" begins with -",
		},
		{
			flag:     "foo=bar",
			errorMsg: "flag \"foo=bar\" contains =",
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("FlagSet.Var(&v, %q, \"\")", test.flag)

		fs := OneCmd("", ContinueOnError)
		errCh := make(chan error)
		go func() {
			defer func() {
				e := recover()
				if err, ok := e.(error); ok {
					errCh <- err
				}
				errCh <- nil
			}()
			var v flagVar
			fs.Var(&v, test.flag, "")
		}()
		var err error
		err = <-errCh
		if msg := test.errorMsg; err != nil && msg != err.Error() {
			t.Errorf("%s\n: unexpected output: expected %q, but got %q", testName, msg, err)
		}
	}
}

func TestRedefinedFlags(t *testing.T) {
	tests := []struct {
		flagSetName string
		errorMsg    string
	}{
		{
			flagSetName: "",
			errorMsg:    "flag redefined: foo",
		},
		{
			flagSetName: "fs",
			errorMsg:    "fs flag redefined: foo",
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("flag redefined in FlagSet(%q)", test.flagSetName)

		fs := OneCmd(test.flagSetName, ContinueOnError)

		var v flagVar
		fs.Var(&v, "foo", "")

		errCh := make(chan error)
		go func() {
			defer func() {
				e := recover()
				if err, ok := e.(error); ok {
					errCh <- err
				}
				errCh <- nil
			}()
			fs.Var(&v, "foo", "")
		}()
		err := <-errCh
		if msg := test.errorMsg; msg != err.Error() {
			t.Errorf("%s\n: unexpected output: expected %q, bug got %q", testName, msg, err)
		}
	}
}
