package structflag

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Set is the minimal interface structflag needs to work.
// It is a subset of flag.FlagSet
type Set interface {
	Args() []string
	BoolVar(p *bool, name string, value bool, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	IntVar(p *int, name string, value int, usage string)
	Parse(arguments []string) error
	PrintDefaults()
	StringVar(p *string, name string, value string, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Var(value flag.Value, name string, usage string)
}

var (
	set Set

	// NewSet  defaults to flag.CommandLine of the standard flag package.
	NewSet = func() Set { return flag.NewFlagSet(os.Args[0], flag.ExitOnError) }
)

var (
	// NameTag is the struct tag used to overwrite
	// the struct field name as flag name.
	// Struct fields with NameTag of "-" will be ignored.
	NameTag = "flag"

	// UsageTag is the struct tag used to give
	// the usage description of a flag
	UsageTag = "usage"

	// DefaultTag is the struct tag used to
	// define the default value for the field
	// (if that default value is different from the zero value)
	DefaultTag = "default"

	// NameFunc is called as last operation for every flag name
	NameFunc = func(name string) string { return name }
)

var (
	flagValueType    = reflect.TypeOf((*flag.Value)(nil)).Elem()
	timeDurationType = reflect.TypeOf(time.Duration(0))
)

// StructVar defines the fields of a struct as flags.
// structPtr must be a pointer to a struct.
// Anonoymous embedded fields are flattened.
// Struct fields with NameTag of "-" will be ignored.
func StructVar(structPtr interface{}) {
	if set == nil {
		set = NewSet()
	}
	structVar(structPtr, set, false)
}

func structVar(structPtr interface{}, set Set, fieldValuesAsDefault bool) {
	var err error
	fields := flatStructFields(reflect.ValueOf(structPtr))
	for _, field := range fields {
		name := field.Tag.Get(NameTag)
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		name = NameFunc(name)

		usage := field.Tag.Get(UsageTag)

		if field.Type.Implements(flagValueType) {
			set.Var(field.Value.Addr().Interface().(flag.Value), name, usage)
			continue
		}

		defaultStr, hasDefault := field.Tag.Lookup(DefaultTag)

		if field.Type == timeDurationType {
			var value time.Duration
			if fieldValuesAsDefault {
				value = field.Value.Interface().(time.Duration)
			} else if hasDefault {
				value, err = time.ParseDuration(defaultStr)
				if err != nil {
					panic(err)
				}
			}
			set.DurationVar(field.Value.Addr().Interface().(*time.Duration), name, value, usage)
			continue
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			var value bool
			if fieldValuesAsDefault {
				value = field.Value.Interface().(bool)
			} else if hasDefault {
				value, err = strconv.ParseBool(defaultStr)
				if err != nil {
					panic(err)
				}
			}
			set.BoolVar(field.Value.Addr().Interface().(*bool), name, value, usage)

		case reflect.Float64:
			var value float64
			if fieldValuesAsDefault {
				value = field.Value.Interface().(float64)
			} else if hasDefault {
				value, err = strconv.ParseFloat(defaultStr, 64)
				if err != nil {
					panic(err)
				}
			}
			set.Float64Var(field.Value.Addr().Interface().(*float64), name, value, usage)

		case reflect.Int64:
			var value int64
			if fieldValuesAsDefault {
				value = field.Value.Interface().(int64)
			} else if hasDefault {
				value, err = strconv.ParseInt(defaultStr, 0, 64)
				if err != nil {
					panic(err)
				}
			}
			set.Int64Var(field.Value.Addr().Interface().(*int64), name, value, usage)

		case reflect.Int:
			var value int64
			if fieldValuesAsDefault {
				value = int64(field.Value.Interface().(int))
			} else if hasDefault {
				value, err = strconv.ParseInt(defaultStr, 0, 64)
				if err != nil {
					panic(err)
				}
			}
			set.IntVar(field.Value.Addr().Interface().(*int), name, int(value), usage)

		case reflect.String:
			var value string
			if fieldValuesAsDefault {
				value = field.Value.Interface().(string)
			} else if hasDefault {
				value = defaultStr
			}
			set.StringVar(field.Value.Addr().Interface().(*string), name, value, usage)

		case reflect.Uint64:
			var value uint64
			if fieldValuesAsDefault {
				value = field.Value.Interface().(uint64)
			} else if hasDefault {
				value, err = strconv.ParseUint(defaultStr, 0, 64)
				if err != nil {
					panic(err)
				}
			}
			set.Uint64Var(field.Value.Addr().Interface().(*uint64), name, value, usage)

		case reflect.Uint:
			var value uint64
			if fieldValuesAsDefault {
				value = uint64(field.Value.Interface().(uint))
			} else if hasDefault {
				value, err = strconv.ParseUint(defaultStr, 0, 64)
				if err != nil {
					panic(err)
				}
			}
			set.UintVar(field.Value.Addr().Interface().(*uint), name, uint(value), usage)
		}
	}
}

// Parse parses args, or if no args are given os.Args[1:]
func Parse(args ...string) ([]string, error) {
	if set == nil {
		set = NewSet()
	}
	return parse(args, set)
}

func parse(args []string, set Set) ([]string, error) {
	if len(args) == 0 {
		args = os.Args[1:]
	}
	err := set.Parse(args)
	if err != nil {
		return nil, err
	}
	return set.Args(), nil
}

// PrintDefaults prints to standard error the default values of all defined command-line flags in the set.
func PrintDefaults() {
	if set == nil {
		set = NewSet()
	}
	set.PrintDefaults()
}

// LoadFileAndParseCommandLine loads the configuration from filename
// into structPtr and then parses the command line.
// Every value that is present in command line overwrites the
// value loaded from the configuration file.
// Values not present in the command line won't effect the Values
// loaded from the configuration file.
// If there is an error loading the configuration file,
// then the command line still gets parsed.
// The error os.ErrNotExist can be ignored if the existence
// of the configuration file is optional.
func LoadFileAndParseCommandLine(filename string, structPtr interface{}) ([]string, error) {
	// Initialize global variable set with unchanged default values
	// so that a later PrintDefaults() prints the correct default values.
	StructVar(structPtr)

	// Load and unmarshal struct from file
	loadErr := LoadFile(filename, structPtr)

	// Use the existing struct values as defaults for tempSet
	// so that not existing args don't overwrite existing values
	// that have been loaded from the confriguration file
	tempSet := NewSet()
	structVar(structPtr, tempSet, true)
	err := tempSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return tempSet.Args(), loadErr
}

// MustLoadFileAndParseCommandLine same as LoadFileAndParseCommandLine but panics on error
func MustLoadFileAndParseCommandLine(filename string, structPtr interface{}) []string {
	args, err := LoadFileAndParseCommandLine(filename, structPtr)
	if err != nil {
		panic(err)
	}
	return args
}

// LoadFileIfExistsAndMustParseCommandLine same as LoadFileAndParseCommandLine but panics on error
func LoadFileIfExistsAndMustParseCommandLine(filename string, structPtr interface{}) []string {
	args, err := LoadFileAndParseCommandLine(filename, structPtr)
	if err != nil && err != os.ErrNotExist {
		panic(err)
	}
	return args
}

type structFieldAndValue struct {
	reflect.StructField
	Value reflect.Value
}

// flatStructFields returns the structFieldAndValue of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func flatStructFields(v reflect.Value) []structFieldAndValue {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	numField := t.NumField()
	fields := make([]structFieldAndValue, 0, numField)
	for i := 0; i < numField; i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if ft.Anonymous {
			fields = append(fields, flatStructFields(fv)...)
		} else {
			fields = append(fields, structFieldAndValue{ft, fv})
		}
	}
	return fields
}

// LoadFile loads a struct from a JSON or XML file.
// The file type is determined by the file extension.
func LoadFile(filename string, structPtr interface{}) error {
	// Load and unmarshal struct from file
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return LoadJSON(filename, structPtr)
	case ".xml":
		return LoadXML(filename, structPtr)
	}
	return errors.New("File extension not supported: " + ext)
}

// LoadXML loads a struct from a XML file
func LoadXML(filename string, structPtr interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, structPtr)
}

// SaveXML saves a struct as a XML file
func SaveXML(filename string, structPtr interface{}, indent ...string) error {
	data, err := xml.MarshalIndent(structPtr, "", strings.Join(indent, ""))
	if err != nil {
		return err
	}
	data = append([]byte(xml.Header), data...)
	return ioutil.WriteFile(filename, data, 0660)
}

// LoadJSON loads a struct from a JSON file
func LoadJSON(filename string, structPtr interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, structPtr)
}

// SaveJSON saves a struct as a JSON file
func SaveJSON(filename string, structPtr interface{}, indent ...string) error {
	data, err := json.MarshalIndent(structPtr, "", strings.Join(indent, ""))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0660)
}

type commandDetails struct {
	command     string
	description string
	action      func([]string) error
}

var (
	// ErrCommandNotFound is returned by Commands.Execute
	// if no matching command could be found.
	ErrCommandNotFound = errors.New("Command not found")

	// ErrNotEnoughArguments is returned when a command
	// is called with not enough aruments
	ErrNotEnoughArguments = errors.New("Not enough argumetns")
)

// Commands helps to parse and execute commands from command line arguments
type Commands []commandDetails

// AddWithArgs adds a command with additional string arguments
func (c *Commands) AddWithArgs(command, description string, action func([]string) error) {
	*c = append(*c, commandDetails{command, description, action})
}

// AddWithArg adds a command with a single additional string argument
func (c *Commands) AddWithArg(command, description string, action func(string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 1 {
			return ErrNotEnoughArguments
		}
		return action(args[0])
	})
}

// AddWith2Args adds a command with two additional string argument
func (c *Commands) AddWith2Args(command, description string, action func(string, string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 2 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1])
	})
}

// AddWith3Args adds a command with threee additional string argument
func (c *Commands) AddWith3Args(command, description string, action func(string, string, string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 3 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1], args[2])
	})
}

// Add adds a command
func (c *Commands) Add(command, description string, action func() error) {
	c.AddWithArgs(command, description, func([]string) error { return action() })
}

// PrintDescription prints a description of all commands to stderr
func (c *Commands) PrintDescription() {
	fmt.Fprint(os.Stderr, "Commands:\n")
	for i, comm := range *c {
		if comm.description == "" {
			fmt.Fprintf(os.Stderr, "%d. %s\n\n", i+1, comm.command)
		} else {
			fmt.Fprintf(os.Stderr, "%d. %s: %s\n\n", i+1, comm.command, comm.description)
		}
	}
	fmt.Fprint(os.Stderr, "Flags:\n")
	PrintDefaults()
}

// Execute executes the command from args[0] or returns
// ErrCommandNotFound if no such command was registered
// of if len(args) == 0
func (c *Commands) Execute(args []string) error {
	if len(args) == 0 {
		return ErrCommandNotFound
	}
	command := args[0]
	for _, comm := range *c {
		if comm.command == command {
			return comm.action(args[1:])
		}
	}
	return ErrCommandNotFound
}
