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
			case reflect.Bool:
				value, err = flags.GetBool(tag)
			case reflect.Int:
				value, err = flags.GetInt(tag)
			case reflect.Int8:
				value, err = flags.GetInt8(tag)
			case reflect.Int16:
				value, err = flags.GetInt16(tag)
			case reflect.Int32:
				value, err = flags.GetInt32(tag)
			case reflect.Int64:
				value, err = flags.GetInt64(tag)
			case reflect.Uint:
				value, err = flags.GetUint(tag)
			case reflect.Uint8:
				value, err = flags.GetUint8(tag)
			case reflect.Uint16:
				value, err = flags.GetUint16(tag)
			case reflect.Uint32:
				value, err = flags.GetUint32(tag)
			case reflect.Uint64:
				value, err = flags.GetUint64(tag)
			case reflect.Float32:
				value, err = flags.GetFloat32(tag)
			case reflect.Float64:
				value, err = flags.GetFloat64(tag)
			case reflect.String:
				value, err = flags.GetString(tag)
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
