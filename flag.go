package flag

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Deprecated: rather call GetDefaultUsage() based on the error you get while calling Parse(args)
// ErrHelp is the error returned if the -help or -h flag is invoked
// but no such flag is defined.
var ErrHelp = errors.New("flag: help requested")

// errParse is returned by Set if a flag's value fails to parse, such as with an invalid integer for Int.
// It then gets wrapped throughmt.Errorf to provide more information.
var errParse = errors.New("parse error")

// errRange is returned by Set if a flag's value is out of range.
// It then gets wrapped throughmt.Errorf to provide more information.
var errRange = errors.New("value out of range")

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return err
}

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
// The flag package may call the String method with a zero-valued receiver,
// such as a nil pointer.
type Value interface {
	String() string
	Set(s string) error
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() any { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

func (b *boolValue) IsBoolFlag() bool { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() any { return int64(*i) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() any { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() any { return uint64(*i) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (v *stringValue) Set(s string) error {
	*v = stringValue(s)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = numError(err)
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() any { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		err = errParse
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() any { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- encoding.TextUnmarshaler Value
type textValue struct{ p encoding.TextUnmarshaler }

func newTextValue(val encoding.TextMarshaler, p encoding.TextUnmarshaler) textValue {
	ptrVal := reflect.ValueOf(p)
	if ptrVal.Kind() != reflect.Ptr {
		panic("variable value type must be a pointer")
	}
	defVal := reflect.ValueOf(val)
	if defVal.Kind() == reflect.Ptr {
		defVal = defVal.Elem()
	}
	if defVal.Type() != ptrVal.Type().Elem() {
		panic(fmt.Sprintf("default type does not match variable type: %v != %v", defVal.Type(), ptrVal.Type().Elem()))
	}
	ptrVal.Elem().Set(defVal)
	return textValue{p}
}

func (v textValue) Set(s string) error {
	return v.p.UnmarshalText([]byte(s))
}

func (v textValue) Get() interface{} {
	return v.p
}

func (v textValue) String() string {
	if m, ok := v.p.(encoding.TextMarshaler); ok {
		if b, err := m.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}

// -- func Value
type funcValue func(string) error

func (f funcValue) Set(s string) error {
	return f(s)
}

func (f funcValue) String() string { return "" }

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface, except the type used by Func.
type Getter interface {
	Value
	Get() any
}

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling int

// These constants cause FlagSet.Parse to behave as described if the parse fails.
const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError                         // Call panic with a descriptive error.
)

// A Flag represents the state of a flag.
type Flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message

	envs     map[string]bool
	cfgs     map[string]bool
	enums    map[string]bool
	alias    map[string]bool
	aliasFor string //this flag is an alias for
}

func isEnumValid(e string, enums []string) bool {
	valid := false
	if len(enums) > 0 {
		for _, s := range enums {
			if e == s {
				valid = true
			}
		}
	} else {
		valid = true
	}
	return valid
}

func (f *Flag) Set(s string) error {
	if !isEnumValid(s, keys(f.enums)) {
		return fmt.Errorf("flag %v is a enum flag, needs one of these values %v", f.Name, strings.Join(keys(f.enums), ", "))
	}
	return f.Value.Set(s)
}

func keys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func qKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, fmt.Sprintf("%q", key))
	}
	return keys
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (f *Command) BoolVar(p *bool, name string, value bool, usage string, features ...*flagFeature) {
	f.Var(newBoolValue(value, p), name, usage, features...)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (f *Command) Bool(name string, value bool, usage string, features ...*flagFeature) *bool {
	p := new(bool)
	f.BoolVar(p, name, value, usage, features...)
	return p
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *Command) IntVar(p *int, name string, value int, usage string, features ...*flagFeature) {
	f.Var(newIntValue(value, p), name, usage, features...)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *Command) Int(name string, value int, usage string, features ...*flagFeature) *int {
	p := new(int)
	f.IntVar(p, name, value, usage, features...)
	return p
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (f *Command) Int64Var(p *int64, name string, value int64, usage string, features ...*flagFeature) {
	f.Var(newInt64Value(value, p), name, usage, features...)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (f *Command) Int64(name string, value int64, usage string, features ...*flagFeature) *int64 {
	p := new(int64)
	f.Int64Var(p, name, value, usage, features...)
	return p
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *Command) UintVar(p *uint, name string, value uint, usage string, features ...*flagFeature) {
	f.Var(newUintValue(value, p), name, usage, features...)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *Command) Uint(name string, value uint, usage string, features ...*flagFeature) *uint {
	p := new(uint)
	f.UintVar(p, name, value, usage, features...)
	return p
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (f *Command) Uint64Var(p *uint64, name string, value uint64, usage string, features ...*flagFeature) {
	f.Var(newUint64Value(value, p), name, usage, features...)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (f *Command) Uint64(name string, value uint64, usage string, features ...*flagFeature) *uint64 {
	p := new(uint64)
	f.Uint64Var(p, name, value, usage, features...)
	return p
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *Command) StringVar(p *string, name string, value string, usage string, features ...*flagFeature) {
	f.Var(newStringValue(value, p), name, usage, features...)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *Command) String(name string, value string, usage string, features ...*flagFeature) *string {
	p := new(string)
	f.StringVar(p, name, value, usage, features...)
	return p
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (f *Command) Float64Var(p *float64, name string, value float64, usage string, features ...*flagFeature) {
	f.Var(newFloat64Value(value, p), name, usage, features...)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (f *Command) Float64(name string, value float64, usage string, features ...*flagFeature) *float64 {
	p := new(float64)
	f.Float64Var(p, name, value, usage, features...)
	return p
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *Command) DurationVar(p *time.Duration, name string, value time.Duration, usage string, features ...*flagFeature) {
	f.Var(newDurationValue(value, p), name, usage, features...)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts a value acceptable to time.ParseDuration.
func (f *Command) Duration(name string, value time.Duration, usage string, features ...*flagFeature) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, value, usage, features...)
	return p
}

// TextVar defines a flag with a specified name, default value, and usage string.
// The argument p must be a pointer to a variable that will hold the value
// of the flag, and p must implement encoding.TextUnmarshaler.
// If the flag is used, the flag value will be passed to p's UnmarshalText method.
// The type of the default value must be the same as the type of p.
func (f *Command) TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string, features ...*flagFeature) {
	f.Var(newTextValue(value, p), name, usage, features...)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (f *Command) Var(value Value, name string, usage string, features ...*flagFeature) {
	_, err := f.varErr(value, name, usage, features...)
	if err != nil {
		panic(err)
	}
}

func (f *Command) varErr(value Value, name string, usage string, features ...*flagFeature) (*Flag, error) {
	// Flag must not begin "-" or contain "=".
	if strings.HasPrefix(name, "-") {
		return nil, fmt.Errorf("flag %q begins with -", name)
	} else if strings.Contains(name, "=") {
		return nil, fmt.Errorf("flag %q contains =", name)
	}

	// Remember the default value as a string; it won't change.
	flag := &Flag{Name: name, Usage: usage, Value: value, DefValue: value.String(), envs: make(map[string]bool), cfgs: make(map[string]bool), enums: make(map[string]bool), alias: make(map[string]bool)}
	_, alreadythere := f.formal[name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = fmt.Sprintf("flag redefined: %s", name)
		} else {
			msg = fmt.Sprintf("%s flag redefined: %s", f.name, name)
		}
		return nil, errors.New(msg) // Happens only if flags are declared with identical names
	}
	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}
	f.formal[name] = flag
	sortedFeatures := flagFeatures(features)
	sort.Sort(sortedFeatures)
	for _, feature := range sortedFeatures {
		feature.add(f, flag)
	}
	return flag, nil
}

type subCommand struct {
	fn func(fs *Command, args []string)
	fs *Command
}

// A Command represents a set of defined flags. The zero value of a Command
// has no name and has ContinueOnError error handling.
//
// Flag names must be unique within a Command. An attempt to define a flag whose
// name is already in use will cause a panic.
type Command struct {
	// Deprecated: rather use GetDefaultUsage() and use the result to show the default usage message
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler. What happens after Usage is called depends
	// on the ErrorHandling setting; for the command line, this defaults
	// to ExitOnError, which exits the program after calling Usage.
	Usage func()

	name          string
	usg           string // description about this command
	parsed        bool
	actual        map[string]*Flag
	formal        map[string]*Flag
	ptrs          map[string]*Flag
	args          []string // arguments after flags
	errorHandling ErrorHandling
	output        io.Writer // Deprecated: nil means stderr; use Output() accessor
	cfgPath       string
	cfg           map[string]interface{}
	SubCmds       map[string]*subCommand
	parentCmd     *Command
}

// sortFlags returns the flags as a slice in lexicographical sorted order.
func sortFlags(flags map[string]*Flag) []*Flag {
	result := make([]*Flag, len(flags))
	i := 0
	for _, f := range flags {
		result[i] = f
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Name returns the name of the flag set.
func (f *Command) Name() string {
	return f.name
}

// ErrorHandling returns the error handling behavior of the flag set.
func (f *Command) ErrorHandling() ErrorHandling {
	return f.errorHandling
}

// Deprecated:
// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (f *Command) SetOutput(output io.Writer) {
	f.output = output
}

// VisitAll visits the flags in lexicographical order, calling fn for each.
// It visits all flags, even those not set.
func (f *Command) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(f.formal) {
		fn(flag)
	}
}

// Visit visits the flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func (f *Command) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(f.actual) {
		fn(flag)
	}
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (f *Command) Lookup(name string) *Flag {
	return f.formal[name]
}

// Set sets the value of the named flag.
func (f *Command) Set(name, value string) error {
	flag, ok := f.formal[name]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	err := flag.Set(value)
	if err != nil {
		return err
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return nil
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(flag *Flag, value string) (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(flag.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, flag.Name, e)
		}
	}()
	return value == z.Interface().(Value).String(), nil
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = valueTypeName(flag.Value)
	return
}

func valueTypeName(v Value) (name string) {
	name = "value"
	switch v.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

// Deprecated: This function is no longer recommended for use.
// use GetDefaultUsage() instead.
// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined command-line flags in the set. See the
// documentation for the global function PrintDefaults for more information.
func (f *Command) PrintDefaults() {
	panic("deprecated")
}

func (f *Command) SetUsage(usage string) {
	f.usg = usage
}

// returns a well formatted short usage to print while user passes help flag
func (f *Command) GetDefaultUsage() (usage string, err error) {
	return f.getDefaultUsage(true)
}

// returns a well formatted detailed (with additional details about the features of the flag) usage to print while user passes help flag
func (f *Command) GetDefaultUsageLong() (usage string, err error) {
	return f.getDefaultUsage(false)
}

// returns a well formatted usage to print while user passes help flag
func (f *Command) getDefaultUsage(short bool) (usage string, err error) {
	var isZeroValueErrs []error
	defaultUsage := ""
	if f.usg != "" {
		defaultUsage += fmt.Sprintf("%s\n\n", f.usg)
	}
	hasFlags := len(f.formal) > 0
	hasSubCmds := len(f.SubCmds) > 0

	currentCmd := f
	commandName := currentCmd.name
	for {
		if currentCmd.parentCmd != nil {
			commandName = currentCmd.parentCmd.name + " " + commandName
			currentCmd = currentCmd.parentCmd
		} else {
			break
		}
	}
	if hasFlags {
		defaultUsage += fmt.Sprintf("usage:\n  %v [<flags>]\n", commandName)
	}
	if hasFlags && hasSubCmds {
		defaultUsage += "  or\n"
	}
	if hasSubCmds {
		defaultUsage += fmt.Sprintf("  %v [<sub-command>]\n", commandName)
	}
	// list of subcommands
	if len(f.SubCmds) > 0 {
		defaultUsage += "\n"
		defaultUsage += "Available sub commands:\n"
		for _, sc := range f.SubCmds {
			defaultUsage += ("  " + sc.fs.name + "  " + sc.fs.usg + "\n")
		}
	}
	if hasFlags {
		defaultUsage += "\nFlags:\n"
		for _, flag := range f.formal {
			usage := flag.Usage
			if flag.aliasFor == "" {
				if usage == "" {
					usage = "usage not available"
				}
				bracketUsage := fmt.Sprintf("defaults to \"%v\"", flag.DefValue)
				if !short {
					if len(flag.enums) > 0 {
						bracketUsage += fmt.Sprintf(", possible values [%v]", strings.Join(qKeys(flag.enums), ", "))
					}
					if len(flag.alias) > 0 {
						bracketUsage += fmt.Sprintf(", alias [%v]", strings.Join(qKeys(flag.alias), ", "))
					}
					if len(flag.envs) > 0 {
						bracketUsage += fmt.Sprintf(", binds to env/s [%v]", strings.Join(qKeys(flag.envs), ", "))
					}
					if len(flag.cfgs) > 0 {
						bracketUsage += fmt.Sprintf(", binds to cfg/s [%v]", strings.Join(qKeys(flag.cfgs), ", "))
					}
				}
				defaultUsage += fmt.Sprintf("  --%v %v  %v, (%v)\n", flag.Name, valueTypeName(flag.Value), usage, bracketUsage)
			} else {
				defaultUsage += fmt.Sprintf("  --%v %v  alias for \"--%v\"\n", flag.Name, valueTypeName(flag.Value), flag.aliasFor)
			}
			if hasSubCmds {
				defaultUsage += fmt.Sprintf("\nUse \"%v [command] --help\" for more information about a command.", commandName)
			}
		}
	}
	errS := ""
	if errs := isZeroValueErrs; len(errs) > 0 {
		for _, err := range errs {
			errS += (err.Error() + "\n")
		}
	}
	if errS != "" {
		err = errors.New(errS)
	}
	return defaultUsage, err
}

// NFlag returns the number of flags that have been set.
func (f *Command) NFlag() int { return len(f.actual) }

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func (f *Command) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *Command) NArg() int { return len(f.args) }

// Args returns the non-flag arguments.
func (f *Command) Args() []string { return f.args }

// Func defines a flag with the specified name and usage string.
// Each time the flag is seen, fn is called with the value of the flag.
// If fn returns a non-nil error, it will be treated as a flag value parsing error.
func (f *Command) Func(name, usage string, fn func(string) error, features ...*flagFeature) {
	f.Var(funcValue(fn), name, usage, features...)
}

// parseOne parses one flag. It reports whether a flag was seen.
func (f *Command) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}
	s := f.args[0]
	if len(s) < 2 || s[0] != '-' {
		return false, nil
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 { // "--" terminates the flags
			f.args = f.args[1:]
			return false, nil
		}
	}
	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return false, fmt.Errorf("bad flag syntax: %s", s)
	}

	// it's a flag. does it have an argument?
	f.args = f.args[1:]
	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}
	m := f.formal
	flag, alreadythere := m[name] // BUG
	if !alreadythere {
		return false, fmt.Errorf("flag provided but not defined: -%s", name)
	}

	if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() { // special case: doesn't need an arg
		if hasValue {
			if err := fv.Set(value); err != nil {
				return false, fmt.Errorf("invalid boolean value %q for -%s: %v", value, name, err)
			}
		} else {
			if err := fv.Set("true"); err != nil {
				return false, fmt.Errorf("invalid boolean flag %s: %v", name, err)
			}
		}
	} else {
		// It must have a value, which might be the next argument.
		if !hasValue && len(f.args) > 0 {
			// value is the next arg
			hasValue = true
			value, f.args = f.args[0], f.args[1:]
		}
		if !hasValue {
			return false, fmt.Errorf("flag needs an argument: -%s", name)
		}
		if err := flag.Set(value); err != nil {
			return false, fmt.Errorf("invalid value %q for flag -%s: %v", value, name, err)
		}
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return true, nil
}

// ParseWithoutArgs parses everything like binding cfg, binding env, binding to other flags etc but arguments passed to the
// program won't be parsed and considered, when you require flag set to act like config loader (viper'ish)
// still takes in arguments to parse the sub commands passed and run it
func (f *Command) ParseWithoutArgs(args []string) error {
	// it is possible that user is trying run a sub-command
	_, err := f.parseSubCommandAndRun(args)
	return err
}

// lets us know whether subcommand found in args and ran
func (f *Command) parseSubCommandAndRun(args []string) (bool, error) {
	SubCmdFsName, SubCmdFsArgs, ok := GetFirstSubCommandWithArgs(args)
	if ok {
		sc, ok := f.SubCmds[SubCmdFsName]
		if !ok {
			return false, fmt.Errorf("you are trying to run subcommand with name %v but it doesn't exist", SubCmdFsName)
		}
		sc.fn(sc.fs, SubCmdFsArgs)
	}
	return ok, nil
}

func GetFirstSubCommandWithArgs(args []string) (string, []string, bool) {
	if len(args) == 0 {
		return "", nil, false
	}
	str := args[0]
	if !strings.Contains(str, "-") {
		return str, args[1:], true
	}
	return "", nil, false
}

func (f *Command) handleError(err error) error {
	switch f.errorHandling {
	case ContinueOnError:
		return err
	case ExitOnError:
		if err == ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	case PanicOnError:
		panic(err)
	}
	return nil
}

// Parse parses everything like binding cfg, binding env, binding to other flags etc
// then parses the arguments and any argument found will override any prior mentioned bindings.
// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *Command) Parse(arguments []string) error {
	ran, err := f.parseSubCommandAndRun(arguments)
	if err != nil {
		return f.handleError(err)
	}
	// did we find a sub command and ran it?
	if ran {
		// then we shouldn't continue running the parent command
		return nil
	}
	f.parsed = true
	f.args = arguments
	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		return f.handleError(err)
	}
	return nil
}

// Parsed reports whether f.Parse has been called.
func (f *Command) Parsed() bool {
	return f.parsed
}

// Cmd is just a interface for *flag.FlagSet
type Cmd interface {

	// ParseWithoutArgs parses the command-line arguments without consuming any of them.
	// It returns an error if there are any unparsed flags or any error encountered during flag parsing.
	ParseWithoutArgs(args []string) error

	// loads all environment variables from a env file using os.SetEnv
	// effectively making it easier to bind the flags to a env
	LoadEnv(path string) error

	// loads a configuration file at path to this command so you can bind configurations
	LoadCfg(path string) (err error)

	// introduces a subcommand to this command
	// you can pass a callback which will recieve a new CMD with name name and args you should parse with the CMD
	// you recieved after defining the flags
	SubCmd(name string, usage string, onCmd func(subCmd Cmd, args []string))

	// add the values possible for the flag you are defining
	Enum(enums ...string) *flagFeature

	// alias for the fla you are defining, like h for a flag named help
	Alias(flags ...string) *flagFeature

	// bind the cfg value from the configurtion file you loaded to the flag you are defining
	Cfg(cfgs ...string) *flagFeature

	//bind env to the flag you are defining
	Env(envs ...string) *flagFeature

	// BoolVar defines a bool flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a bool variable in which to store the value of the flag.
	BoolVar(p *bool, name string, value bool, usage string, features ...*flagFeature)

	// Bool defines a bool flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a bool variable that stores the value of the flag.
	Bool(name string, value bool, usage string, features ...*flagFeature) *bool

	// IntVar defines an int flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to an int variable in which to store the value of the flag.
	IntVar(p *int, name string, value int, usage string, features ...*flagFeature)

	// Int defines an int flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of an int variable that stores the value of the flag.
	Int(name string, value int, usage string, features ...*flagFeature) *int

	// Int64Var defines an int64 flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to an int64 variable in which to store the value of the flag.
	Int64Var(p *int64, name string, value int64, usage string, features ...*flagFeature)

	// Int64 defines an int64 flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of an int64 variable that stores the value of the flag.
	Int64(name string, value int64, usage string, features ...*flagFeature) *int64

	// UintVar defines a uint flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a uint variable in which to store the value of the flag.
	UintVar(p *uint, name string, value uint, usage string, features ...*flagFeature)

	// Uint defines a uint flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a uint variable that stores the value of the flag.
	Uint(name string, value uint, usage string, features ...*flagFeature) *uint

	// Uint64Var defines a uint64 flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a uint64 variable in which to store the value of the flag.
	Uint64Var(p *uint64, name string, value uint64, usage string, features ...*flagFeature)

	// Uint64 defines a uint64 flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a uint64 variable that stores the value of the flag.
	Uint64(name string, value uint64, usage string, features ...*flagFeature) *uint64

	// StringVar defines a string flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a string variable in which to store the value of the flag.
	StringVar(p *string, name string, value string, usage string, features ...*flagFeature)

	// String defines a string flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a string variable that stores the value of the flag.
	String(name string, value string, usage string, features ...*flagFeature) *string

	// Float64Var defines a float64 flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a float64 variable in which to store the value of the flag.
	Float64Var(p *float64, name string, value float64, usage string, features ...*flagFeature)

	// Float64 defines a float64 flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a float64 variable that stores the value of the flag.
	Float64(name string, value float64, usage string, features ...*flagFeature) *float64

	// DurationVar defines a time.Duration flag with specified name, default value, usage string, and optional flag features.
	// The argument p points to a time.Duration variable in which to store the value of the flag.
	DurationVar(p *time.Duration, name string, value time.Duration, usage string, features ...*flagFeature)

	// Duration defines a time.Duration flag with specified name, default value, usage string, and optional flag features.
	// The return value is the address of a time.Duration variable that stores the value of the flag.
	Duration(name string, value time.Duration, usage string, features ...*flagFeature) *time.Duration

	// TextVar defines a flag with specified name, default value, usage string, and optional flag features.
	// The argument p is an encoding.TextUnmarshaler that is used to unmarshal the flag value.
	TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string, features ...*flagFeature)

	// Var defines a flag with specified name, usage string, and optional flag features.
	// The argument value is a Value interface that provides the flag's value and default value.
	Var(value Value, name string, usage string, features ...*flagFeature)

	// Name returns the name of the FlagSet.
	Name() string

	// Set sets the value of the named flag.
	// It returns an error if the flag does not exist or the value is invalid.
	Set(name, value string) error

	// GetDefaultUsage returns the default usage string for the CMD.
	GetDefaultUsage() (usage string, err error)

	// returns a well formatted detailed (with additional details about the features of the flag) usage to print while user passes help flag
	GetDefaultUsageLong() (usage string, err error)

	// Func defines a flag with specified name, usage string, and function to be called when the flag is parsed.
	// The provided function is called with the flag's value as its argument.
	Func(name, usage string, fn func(string) error, features ...*flagFeature)

	// Parse parses the command-line arguments.
	// It returns an error if there are any unparsed flags or any error encountered during flag parsing.
	Parse(arguments []string) error

	// Parsed returns whether the command-line arguments have been parsed.
	Parsed() bool
}

// only the thing you'll ever need
// calls fn with os.args
// see here https://github.com/ondbyte/turbo_flag#alternative
func MainCmd(name string, usage string, errorHandling ErrorHandling, onCmd func(cmd Cmd, args []string)) {
	f := &Command{
		name:          name,
		errorHandling: errorHandling,
		SubCmds:       make(map[string]*subCommand),
		ptrs:          make(map[string]*Flag),
		cfg:           make(map[string]interface{}),
		usg:           usage,
	}
	f.Usage = func() {
		panic("Deprecated")
	}
	onCmd(f, os.Args[1:])
}

// Init sets the name and error handling property for a flag set.
// By default, the zero FlagSet uses an empty name and the
// ContinueOnError error handling policy.
func (f *Command) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}

// loads all environment variables from a env file using os.SetEnv
// effectively making it easier to bind the flags to a env
func (fs *Command) LoadEnv(path string) error {
	if fs.parentCmd != nil {
		return fs.parentCmd.LoadEnv(path)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read env file at %v : %v", path, err)
	}
	envs, err := EnvToMap(string(b))
	if err != nil {
		return fmt.Errorf("failed to parse env file content at %v : %v", path, err)
	}
	errs := ""
	for k, v := range envs {
		err := os.Setenv(k, v)
		if err != nil {
			errs += err.Error() + "\n"
		}
	}
	if errs != "" {
		return fmt.Errorf("partially loaded envs beacuse : %v", err)
	}
	return nil
}

// loads a cfg to this flagset
// any sub command defined will also derive from this
func (fs *Command) LoadCfg(path string) (err error) {
	if fs.parentCmd != nil {
		return fs.parentCmd.LoadCfg(path)
	}
	if path == "" {
		return fmt.Errorf("path is empty while loading config")
	}

	fs.cfgPath = path
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file at %v : %v", path, err)
	}
	fileContent := string(b)
	mapContent := make(map[string]interface{})
	ext := strings.ToUpper(filepath.Ext(path))
	switch ext {
	case "":
		return fmt.Errorf("config file has no extension, add a supported extension [YAML,YML,JSON,PROPERTIES]")
	case ".JSON":
		mapContent, err = JSONToMap(fileContent)
		break
	case ".YML", ".YAML":
		mapContent, err = YAMLToMap(fileContent)
		break
	case ".TOML":
		mapContent, err = TOMLToMap(fileContent)
		break
	default:
		return fmt.Errorf("unsupported extension %v", ext)
	}
	if err != nil {
		return fmt.Errorf("unable to read config file : %v", err)
	}
	fs.cfg = mapContent
	bindCfgRecursiveAfterLoadCfg(fs)
	return nil
}

func bindCfgRecursiveAfterLoadCfg(fs *Command) {
	for _, sc := range fs.SubCmds {
		bindCfgRecursiveAfterLoadCfg(sc.fs)
	}
	for _, flag := range fs.formal {
		fs.bindCfg(flag, keys(flag.cfgs)...)
	}
}

// adds a new sub flagset to the parent flagset, loads the config file if it exists in the parent
// the sub command fn recieves the new FlagSet and the arguments thats for the sub command
// you can add new flags to this sub flagset and call fs.Parse with the arguments you recieved in this function
func (fs *Command) SubCmd(name string, usage string, fn func(cmd Cmd, args []string)) {
	subFs := &Command{name: name, errorHandling: fs.errorHandling}
	subFs.SetUsage(usage)
	//subFs.LoadCfg(fs.cfgPath)
	subFs.cfgPath = fs.cfgPath
	subFs.cfg = fs.cfg
	subFs.parentCmd = fs
	fs.SubCmds[name] = &subCommand{
		fn: func(fs *Command, args []string) {
			var c Cmd
			c = fs
			fn(c, args)
		},
		fs: subFs,
	}
}

type flagFeature struct {
	index int
	add   func(fs *Command, fflag *Flag)
}

type flagFeatures []*flagFeature

func (s flagFeatures) Len() int           { return len(s) }
func (s flagFeatures) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s flagFeatures) Less(i, j int) bool { return s[i].index < s[j].index }

// bind enums to the flag, if you do this only entries in the enums will be the possible values for flag you are currently defining
// https://github.com/ondbyte/turbo_flag#setting-enumsoptionsallowed-values-for-a-flag
func (fs *Command) Enum(enums ...string) *flagFeature {
	return &flagFeature{
		index: 10,
		add: func(fs *Command, f *Flag) {
			fs.bindEnum(f, enums...)
		},
	}
}

func (fs *Command) bindEnum(to *Flag, enums ...string) {
	if !isEnumValid(to.DefValue, enums) {
		panic(fmt.Errorf("you are trying to add enum feature to flag name [%v] but the default value of the flag is %v, default value should be one of the value from enums %v", to.Name, to.DefValue, enums))
	}
	for _, enum := range enums {
		to.enums[enum] = true
	}
}

// binds flag/s with names to the flag you are defining, useful in adding short flags (bind flag help to flag h),
// every property of the flag will be copied.
// https://github.com/ondbyte/turbo_flag#setting-alias-for-a-flag
func (fs *Command) Alias(flags ...string) *flagFeature {
	return &flagFeature{
		index: 11,
		add: func(fs *Command, f *Flag) {
			fs.alias(f, flags...)
		},
	}
}
func (fs *Command) alias(to *Flag, names ...string) {
	for _, name := range names {
		if name == to.Name {
			panic(fmt.Sprintf("cannot add alias to the flag with the same name %v", name))
		}
		f, err := fs.varErr(to.Value, name, to.Usage)
		if err != nil {
			panic(fmt.Sprintf("error while adding enums to the flag %v to flag %v : %v", name, to.Name, err))
		}
		f.envs = to.envs
		f.cfgs = to.cfgs
		f.enums = to.enums
		f.aliasFor = to.Name
		for k, v := range to.alias {
			f.alias[k] = v
		}
		to.alias[f.Name] = true
		f.alias[to.Name] = true
	}
}

// binds configurations value from config file to the to flag,
// use dot notation of the config key to bind.
// https://github.com/ondbyte/turbo_flag#loading-configurations
func (fs *Command) Cfg(cfgs ...string) *flagFeature {
	return &flagFeature{
		index: 8,
		add: func(fs *Command, f *Flag) {
			fs.bindCfg(f, cfgs...)
		},
	}
}

func (fs *Command) bindCfg(to *Flag, cfgs ...string) {
	for _, notation := range cfgs {
		val, err := getValueByDotNotation(fs.cfg, notation)
		if err == nil && val != "" {
			err := to.Set(val)
			if err != nil {
				panic(fmt.Errorf("unable to set notation %v value %v to flag %v", notation, val, to.Name))
			}
		} else {
			cs, err := setValueByDotNotation(fs.cfg, notation, to.Value.String())
			if err == nil {
				for k, v := range cs {
					fs.cfg[k] = v
				}
			}
		}
	}
	for _, cfg := range cfgs {
		to.cfgs[cfg] = true
	}
}

// binds env/s to the to flag you are defining
// https://github.com/ondbyte/turbo_flag#binding-environment-variables
func (fs *Command) Env(envs ...string) *flagFeature {
	return &flagFeature{
		index: 7,
		add: func(fs *Command, f *Flag) {
			fs.bindEnv(f, envs...)
		},
	}
}

func (fs *Command) bindEnv(to *Flag, envs ...string) {
	for _, env := range envs {
		to.envs[env] = true
		val := os.Getenv(env)
		if val != "" {
			err := to.Set(val)
			if err != nil {
				panic(fmt.Errorf("error while setting value from environment, flag name %v,env %v,value %v : %v", to.Name, env, val, err))
			}
		}
	}
}

func (fs *Command) GetFlagForPtr(ptr interface{}) (*Flag, error) {
	key := fmt.Sprint(&ptr)
	ff := fs.ptrs[key]
	if ff != nil {
		return ff, nil
	}
	key = fmt.Sprint(ptr)
	ff = fs.ptrs[key]
	if ff != nil {
		return ff, nil
	}
	return nil, fmt.Errorf(fmt.Sprint(
		"you need to pass the pointer of the flag variable, ",
		"for example you are creating a int var using ",
		"\n",
		`port:=cmd.Int("port","5555","")`,
		"\n",
		"//port is a pointer",
		`cmd.Bind<Cfg/Env/Enum>(port,...)`,
		"\n",
		`pin:=0`,
		"\n",
		`cmd.IntVar(&pin,"pin",12345,"")`,
		"\n",
		"//pin is a variable\n",
		`cmd.Bind<Cfg/Env/Enum>(&pin,...)`,
	))
}

// use this for just a single command
func OneCmd(name string, errorHandling ErrorHandling) Cmd {
	fs := &Command{name: name, errorHandling: errorHandling}
	fs.Usage = func() {
		panic("Deprecated")
	}
	return fs
}
