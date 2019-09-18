package tools

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/logger"
)

// Flag describe a flag
type Flag struct {
	prefix       string
	docPrefix    string
	name         string
	label        string
	defaultValue interface{}
}

// NewFlag creates new instance of Flag
func NewFlag(prefix, docPrefix string) *Flag {
	docPrefixValue := prefix
	if prefix == "" {
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
func (f *Flag) Default(defaultValue interface{}) *Flag {
	f.defaultValue = defaultValue

	return f
}

// Label defines label of Flag
func (f *Flag) Label(label string) *Flag {
	f.label = label

	return f
}

// ToString build Flag Set for string
func (f *Flag) ToString(fs *flag.FlagSet) *string {
	name := fmt.Sprintf("%s%s", f.prefix, f.name)
	envName := strings.ToUpper(SnakeCase(fmt.Sprintf("%s%s", fs.Name(), name)))

	return fs.String(FirstLowerCase(name), LookupEnvString(envName, f.defaultValue.(string)), f.formatLabel(envName))
}

// ToInt build Flag Set for int
func (f *Flag) ToInt(fs *flag.FlagSet) *int {
	name := fmt.Sprintf("%s%s", f.prefix, f.name)
	envName := strings.ToUpper(SnakeCase(fmt.Sprintf("%s%s", fs.Name(), name)))

	return fs.Int(FirstLowerCase(name), LookupEnvInt(envName, f.defaultValue.(int)), f.formatLabel(envName))
}

// ToUint build Flag Set for uint
func (f *Flag) ToUint(fs *flag.FlagSet) *uint {
	name := fmt.Sprintf("%s%s", f.prefix, f.name)
	envName := strings.ToUpper(SnakeCase(fmt.Sprintf("%s%s", fs.Name(), name)))

	return fs.Uint(FirstLowerCase(name), LookupEnvUint(envName, f.defaultValue.(uint)), f.formatLabel(envName))
}

// ToBool build Flag Set for bool
func (f *Flag) ToBool(fs *flag.FlagSet) *bool {
	name := fmt.Sprintf("%s%s", f.prefix, f.name)
	envName := strings.ToUpper(SnakeCase(fmt.Sprintf("%s%s", fs.Name(), name)))

	return fs.Bool(FirstLowerCase(name), LookupEnvBool(envName, f.defaultValue.(bool)), f.formatLabel(envName))
}

func (f *Flag) formatLabel(envName string) string {
	return fmt.Sprintf("[%s] %s {%s}", f.docPrefix, f.label, envName)
}

// LookupEnvString search for given key in environment
func LookupEnvString(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}

// LookupEnvInt search for given key in environment as int
func LookupEnvInt(key string, defaultValue int) int {
	val, ok := os.LookupEnv(key)

	if !ok {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err == nil {
		return intVal
	}

	logger.Warn("%s=%s, not a valid integer: %s", key, val, err)

	return defaultValue
}

// LookupEnvUint search for given key in environment as uint
func LookupEnvUint(key string, defaultValue uint) uint {
	val, ok := os.LookupEnv(key)

	if !ok {
		return defaultValue
	}

	intVal, err := strconv.ParseUint(val, 10, 32)
	if err == nil {
		return uint(intVal)
	}

	logger.Warn("%s=%s, not a valid unsigned integer: %s", key, val, err)

	return defaultValue
}

// LookupEnvBool search for given key in environment as bool
func LookupEnvBool(key string, defaultValue bool) bool {
	val, ok := os.LookupEnv(key)

	if !ok {
		return defaultValue
	}

	boolBal, err := strconv.ParseBool(val)
	if err == nil {
		return boolBal
	}

	logger.Warn("%s=%s, not a valid boolean: %s", key, val, err)

	return defaultValue
}
