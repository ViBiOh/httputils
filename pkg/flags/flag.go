package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Override is an override of default value
type Override struct {
	value interface{}
	name  string
}

// NewOverride create a default override value
func NewOverride(name string, value interface{}) Override {
	return Override{
		name:  name,
		value: value,
	}
}

// Default returns default value for param's name based on default value and overrides
func Default(name string, value interface{}, overrides []Override) interface{} {
	for _, override := range overrides {
		if strings.EqualFold(name, override.name) {
			return override.value
		}
	}

	return value
}

// Flag describe a flag
type Flag struct {
	value     interface{}
	prefix    string
	docPrefix string
	name      string
	label     string
}

// New creates new instance of Flag
func New(prefix, docPrefix string) *Flag {
	docPrefixValue := prefix
	if len(prefix) == 0 {
		docPrefixValue = docPrefix
	}

	return &Flag{
		prefix:    FirstUpperCase(prefix),
		docPrefix: docPrefixValue,
	}
}

// Name defines name of Flag
func (f *Flag) Name(name string) *Flag {
	f.name = name

	return f
}

// Default defines default value of Flag
func (f *Flag) Default(value interface{}) *Flag {
	f.value = value

	return f
}

// Label defines label of Flag
func (f *Flag) Label(format string, a ...interface{}) *Flag {
	f.label = fmt.Sprintf(format, a...)

	return f
}

// ToString build Flag Set for string
func (f *Flag) ToString(fs *flag.FlagSet) *string {
	name, envName := f.getNameAndEnv(fs)
	return fs.String(FirstLowerCase(name), LookupEnvString(envName, f.value.(string)), f.formatLabel(envName))
}

// ToInt build Flag Set for int
func (f *Flag) ToInt(fs *flag.FlagSet) *int {
	name, envName := f.getNameAndEnv(fs)
	return fs.Int(FirstLowerCase(name), LookupEnvInt(envName, f.value.(int)), f.formatLabel(envName))
}

// ToUint build Flag Set for uint
func (f *Flag) ToUint(fs *flag.FlagSet) *uint {
	name, envName := f.getNameAndEnv(fs)

	var value uint
	switch f.value.(type) {
	case int:
		value = uint(f.value.(int))
	case uint:
		value = f.value.(uint)
	default:
		value = 0
	}

	return fs.Uint(FirstLowerCase(name), LookupEnvUint(envName, value), f.formatLabel(envName))
}

// ToFloat64 build Flag Set for float64
func (f *Flag) ToFloat64(fs *flag.FlagSet) *float64 {
	name, envName := f.getNameAndEnv(fs)
	return fs.Float64(FirstLowerCase(name), LookupEnvFloat64(envName, f.value.(float64)), f.formatLabel(envName))
}

// ToBool build Flag Set for bool
func (f *Flag) ToBool(fs *flag.FlagSet) *bool {
	name, envName := f.getNameAndEnv(fs)
	return fs.Bool(FirstLowerCase(name), LookupEnvBool(envName, f.value.(bool)), f.formatLabel(envName))
}

func (f *Flag) getNameAndEnv(fs *flag.FlagSet) (string, string) {
	name := fmt.Sprintf("%s%s", f.prefix, FirstUpperCase(f.name))
	return name, strings.ToUpper(SnakeCase(fmt.Sprintf("%s%s", FirstUpperCase(fs.Name()), FirstUpperCase(name))))
}

func (f *Flag) formatLabel(envName string) string {
	return fmt.Sprintf("[%s] %s {%s}", f.docPrefix, f.label, envName)
}

// LookupEnvString search for given key in environment
func LookupEnvString(key, value string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return value
}

// LookupEnvInt search for given key in environment as int
func LookupEnvInt(key string, value int) int {
	val, ok := os.LookupEnv(key)

	if !ok {
		return value
	}

	intVal, err := strconv.Atoi(val)
	if err == nil {
		return intVal
	}

	return value
}

// LookupEnvUint search for given key in environment as uint
func LookupEnvUint(key string, value uint) uint {
	val, ok := os.LookupEnv(key)

	if !ok {
		return value
	}

	intVal, err := strconv.ParseUint(val, 10, 32)
	if err == nil {
		return uint(intVal)
	}

	return value
}

// LookupEnvFloat64 search for given key in environment as float64
func LookupEnvFloat64(key string, value float64) float64 {
	val, ok := os.LookupEnv(key)

	if !ok {
		return value
	}

	intVal, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return intVal
	}

	return value
}

// LookupEnvBool search for given key in environment as bool
func LookupEnvBool(key string, value bool) bool {
	val, ok := os.LookupEnv(key)

	if !ok {
		return value
	}

	boolBal, err := strconv.ParseBool(val)
	if err == nil {
		return boolBal
	}

	return value
}
