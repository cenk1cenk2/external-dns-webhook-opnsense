package services

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Instance *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if t := strings.SplitN(fld.Tag.Get("json"), ",", 2); len(t) >= 1 {
			return t[0]
		} else if t := fld.Tag.Get("param"); t != "" {
			return fmt.Sprintf("[Route parameter: %s]", t)
		}

		return fld.Name
	})

	return &Validator{
		Instance: v,
	}
}

func (v *Validator) Validate(i any) error {
	rv := reflect.ValueOf(i)

	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)

			for elem.Kind() == reflect.Pointer {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}

			if elem.Kind() == reflect.Struct {
				var ptr any
				if elem.CanAddr() {
					ptr = elem.Addr().Interface()
				} else {
					return fmt.Errorf("validate requires addressable struct elements, element %d is not addressable", i)
				}

				if err := defaults.Set(ptr); err != nil {
					return fmt.Errorf("can not set defaults for element %d: %w", i, err)
				}

				if err := v.Instance.Struct(ptr); errors.Is(err, validator.ValidationErrors{}) {
					return v.format(err.(validator.ValidationErrors))
				} else if err != nil {
					return fmt.Errorf("can not validate element %d: %w", i, err)
				}
			}
		}
		return nil
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	var ptr any
	if rv.CanAddr() {
		ptr = rv.Addr().Interface()
	} else {
		return fmt.Errorf("validate requires an addressable struct (pointer), got non-addressable %T", i)
	}

	if err := defaults.Set(ptr); err != nil {
		return fmt.Errorf("can not set defaults: %w", err)
	}

	if err := v.Instance.Struct(ptr); errors.Is(err, validator.ValidationErrors{}) {
		return v.format(err.(validator.ValidationErrors))
	} else if err != nil {
		return fmt.Errorf("can not validate: %w", err)
	}

	return nil
}

func (v *Validator) format(validationErrs validator.ValidationErrors) error {
	errs := []string{}
	for _, err := range validationErrs {
		e := fmt.Sprintf(
			`"%s" field failed validation: %s`,
			err.Namespace(),
			err.Tag(),
		)

		param := err.Param()
		if param != "" {
			e = fmt.Sprintf("%s -> %s", e, param)
		}

		errs = append(errs, e)
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errs, " | "))
}
