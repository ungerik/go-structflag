package structflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"
)

// Flags is the minimal interface structflag needs to work.
// It is a subset of flag.FlagSet
type Flags interface {
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
	// Output used for printing usage
	Output io.Writer = os.Stderr

	// AppName is the name of the application, defaults to os.Args[0]
	AppName = os.Args[0]

	// OnParseError defines the behaviour if there is an
	// error while parsing the flags.
	// See https://golang.org/pkg/flag/#ErrorHandling
	OnParseError = flag.ExitOnError

	// NewFlags returns new Flags, defaults to flag.NewFlagSet(AppName, OnParseError).
	NewFlags = func() Flags {
		flagSet := flag.NewFlagSet(AppName, OnParseError)
		flagSet.Usage = PrintUsage
		return flagSet
	}

	flags Flags
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

func getOrCreateFlags() Flags {
	if flags == nil {
		flags = NewFlags()
	}
	return flags
}

// StructVar defines the fields of a struct as flags.
// structPtr must be a pointer to a struct.
// Anonoymous embedded fields are flattened.
// Struct fields with NameTag of "-" will be ignored.
func StructVar(structPtr interface{}) {
	structVar(structPtr, getOrCreateFlags(), false)
}

func structVar(structPtr interface{}, flags Flags, fieldValuesAsDefault bool) {
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
			flags.Var(field.Value.Addr().Interface().(flag.Value), name, usage)
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
			flags.DurationVar(field.Value.Addr().Interface().(*time.Duration), name, value, usage)
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
			flags.BoolVar(field.Value.Addr().Interface().(*bool), name, value, usage)

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
			flags.Float64Var(field.Value.Addr().Interface().(*float64), name, value, usage)

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
			flags.Int64Var(field.Value.Addr().Interface().(*int64), name, value, usage)

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
			flags.IntVar(field.Value.Addr().Interface().(*int), name, int(value), usage)

		case reflect.String:
			var value string
			if fieldValuesAsDefault {
				value = field.Value.Interface().(string)
			} else if hasDefault {
				value = defaultStr
			}
			flags.StringVar(field.Value.Addr().Interface().(*string), name, value, usage)

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
			flags.Uint64Var(field.Value.Addr().Interface().(*uint64), name, value, usage)

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
			flags.UintVar(field.Value.Addr().Interface().(*uint), name, uint(value), usage)
		}
	}
}

// Parse parses args, or if no args are given os.Args[1:]
func Parse(args ...string) ([]string, error) {
	return parse(args, getOrCreateFlags())
}

func parse(args []string, flags Flags) ([]string, error) {
	if len(args) == 0 {
		args = os.Args[1:]
	}
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}
	return flags.Args(), nil
}

// PrintUsageTo prints a description of all commands and flags of Set and Commands to output
func PrintUsageTo(output io.Writer) {
	if len(Commands) > 0 {
		fmt.Fprint(Output, "Commands:\n")
		Commands.PrintUsage()
		if flags != nil {
			fmt.Fprint(Output, "Flags:\n")
		}
	}
	if flags != nil {
		flags.PrintDefaults()
	}
}

// PrintUsage prints a description of all commands and flags of Set and Commands to Output
func PrintUsage() {
	PrintUsageTo(Output)
}

// LoadFileAndParseCommandLine loads the configuration from filename
// into structPtr and then parses the command line.
// Every value that is present in command line overwrites the
// value loaded from the configuration file.
// Values not present in the command line won't effect the Values
// loaded from the configuration file.
// If there is an error loading the configuration file,
// then the command line still gets parsed.
// An error where os.IsNotExist(err) == true can be ignored
// if the existence of the configuration file is optional.
func LoadFileAndParseCommandLine(filename string, structPtr interface{}) ([]string, error) {
	// Initialize global variable set with unchanged default values
	// so that a later PrintDefaults() prints the correct default values.
	StructVar(structPtr)

	// Load and unmarshal struct from file
	loadErr := LoadFile(filename, structPtr)

	// Use the existing struct values as defaults for tempSet
	// so that not existing args don't overwrite existing values
	// that have been loaded from the confriguration file
	tempFlags := NewFlags()
	structVar(structPtr, tempFlags, true)
	err := tempFlags.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}
	return tempFlags.Args(), loadErr
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

// PrintConfig prints the flattened struct fields from structPtr to Output.
func PrintConfig(structPtr interface{}) {
	for _, field := range flatStructFields(reflect.ValueOf(structPtr)) {
		fmt.Fprintf(Output, "%s: %v\n", field.Name, field.Value.Interface())
	}
}
