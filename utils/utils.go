package utils

import (
	"bookwise/utils/validator"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

type Envelope map[string]any

// func ConverterByTag[T any](src any, tagName string) (*T, error) {
// 	dst := new(T)
// 	srcVal := reflect.ValueOf(src)
// 	if srcVal.Kind() == reflect.Pointer {
// 		srcVal = srcVal.Elem()
// 	}

// 	dstVal := reflect.ValueOf(dst).Elem()
// 	dstType := dstVal.Type()

// 	for i := 0; i < dstVal.NumField(); i++ {
// 		dstField := dstType.Field(i)
// 		srcFieldName := dstField.Tag.Get(tagName)

// 		if srcFieldName == "" {
// 			srcFieldName = dstField.Name
// 		}

// 		srcField := srcVal.FieldByName(srcFieldName)

// 		if srcField.IsValid() && srcField.Type().AssignableTo(dstField.Type) {
// 			dstVal.Field(i).Set(srcField)
// 		}
// 	}

// 	return dst, nil
// }

func ConverterByTag[T any](src any, tagName string) (*T, error) {
	dst := new(T)
	err := convert(reflect.ValueOf(src), reflect.ValueOf(dst), tagName)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

func convert(srcVal, dstVal reflect.Value, tag string) error {
	// unwrap ponteiros
	for srcVal.Kind() == reflect.Pointer {
		if srcVal.IsNil() {
			return nil
		}
		srcVal = srcVal.Elem()
	}

	for dstVal.Kind() == reflect.Pointer {
		dstVal.Set(reflect.New(dstVal.Type().Elem()))
		dstVal = dstVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return nil
	}

	dstType := dstVal.Type()

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstType.Field(i)
		dstFieldVal := dstVal.Field(i)

		if !dstFieldVal.CanSet() {
			continue
		}

		srcFieldName := dstField.Tag.Get(tag)
		if srcFieldName == "" {
			srcFieldName = dstField.Name
		}

		srcFieldVal := srcVal.FieldByName(srcFieldName)
		if !srcFieldVal.IsValid() {
			continue
		}

		// unwrap ponteiro da origem
		srcField := srcFieldVal
		for srcField.Kind() == reflect.Pointer {
			if srcField.IsNil() {
				continue
			}
			srcField = srcField.Elem()
		}

		// 1️⃣ tipos simples
		if srcField.Type().AssignableTo(dstFieldVal.Type()) {
			dstFieldVal.Set(srcField)
			continue
		}

		// 2️⃣ struct → recursivo
		if srcField.Kind() == reflect.Struct &&
			dstFieldVal.Kind() == reflect.Struct {
			_ = convert(srcField, dstFieldVal, tag)
			continue
		}

		// 3️⃣ *struct → *struct
		if srcField.Kind() == reflect.Struct &&
			dstFieldVal.Kind() == reflect.Pointer &&
			dstFieldVal.Type().Elem().Kind() == reflect.Struct {

			dstFieldVal.Set(reflect.New(dstFieldVal.Type().Elem()))
			_ = convert(srcField, dstFieldVal.Elem(), tag)
		}
	}

	return nil
}

func ReadIntPathVariable(r *http.Request, key string) (int64, error) {
	s := chi.URLParam(r, key)

	if s == "" {
		return 0, fmt.Errorf("missing path parameter: %s", key)
	}

	value, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}

	return value, nil
}

func MinifySQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func ReadString(qs url.Values, key, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}
	return s
}

func ReadInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

func ReadJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func GetTypeName(v any) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return strings.ToLower(t.Name())
}

func WriteJSON(w http.ResponseWriter, status int, data Envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func GenerateRandomCode() int {
	return rand.Intn(900000) + 100000
}

func RunInTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	fnErr := fn(tx)

	if fnErr == nil {
		return tx.Commit()
	}

	if rbErr := tx.Rollback(); rbErr != nil {
		return errors.Join(fnErr, rbErr)
	}

	return fnErr
}
