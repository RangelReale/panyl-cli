package panylcli

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
)

func ParseFlags(flags *pflag.FlagSet, out interface{}) error {
	structType := reflect.TypeOf(out)
	if structType.Kind() != reflect.Ptr {
		return fmt.Errorf("must pass a pointer")
	}
	structVal := structType.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("must pass a struct")
	}

	structPointer := reflect.ValueOf(out).Elem()
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		tag := field.Tag.Get("flag")
		if tag != "" {
			fieldPointer := structPointer.FieldByName(field.Name)
			if !fieldPointer.CanSet() {
				continue
			}

			var value interface{}
			var err error

			switch field.Type.Kind() {
			case reflect.String:
				value, err = flags.GetString(tag)
			case reflect.Int:
				value, err = flags.GetInt(tag)
			default:
				err = fmt.Errorf("unsupported flag type '%s'", field.Type.Name())
			}
			if err != nil {
				return err
			}
			fieldPointer.Set(reflect.ValueOf(value))
		}
	}

	return nil
}
