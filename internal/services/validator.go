package services

import (
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

func (v *Validator) Validate(i interface{}) error {
	if err := defaults.Set(i); err != nil {
		return fmt.Errorf("Can not set defaults: %w", err)
	}

	err := v.Instance.Struct(i)

	if err != nil {
		errs := []string{}
		for _, err := range err.(validator.ValidationErrors) {
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

		return fmt.Errorf("Validation failed: %s", strings.Join(errs, " | "))
	}

	return nil
}
