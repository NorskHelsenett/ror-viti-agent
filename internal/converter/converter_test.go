package converter_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/NorskHelsenett/ror-viti-agent/internal/converter"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
	"github.com/NorskHelsenett/ror/pkg/rorresources/rortypes"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/vitistack/common/pkg/v1alpha1"
)

func TestConvertToVitiMachine(t *testing.T) {
	rorresource := rorresources.NewRorResource("machine", "machine.ror.internal/v1alpha1")
	machine := &rortypes.ResourceMachine{}
	gofakeit.Struct(&machine)
	rorresource.SetMachine(machine)

	vitiresource, err := converter.ConvertToVitiMachine(*rorresource)
	if err != nil {
		t.Error("received unexpected error. %w", err)
	}
	if vitiresource == nil {
		t.Error("received nil on output")
	}

	err = isEmpty(vitiresource.Status)
	if err != nil {
		t.Error("checking status is empty failed. %w", err)
	}
}

func TestConvertToRorMachine(t *testing.T) {
	vitiresource := v1alpha1.Machine{}
	gofakeit.Struct(&vitiresource)

	rorresource, err := converter.ConvertToRorMachine(&vitiresource)
	if err != nil {
		t.Error("received unexpected error. %w", err)
	}
	if rorresource == nil {
		t.Error("received nil on output")
	}
	resource := rorresource.Machine().Get()

	if resource == nil {
		t.Error("received nil on output")
	}
}

func isEmpty(s any) error {
	// first make sure that the input is a struct
	// having any other type, especially a pointer to a struct,
	// might result in panic
	structType := reflect.TypeOf(s)
	if structType.Kind() != reflect.Struct {
		return errors.New("input param should be a struct")
	}

	// now go one by one through the fields and validate their value
	structVal := reflect.ValueOf(s)
	fieldNum := structVal.NumField()

	for i := range fieldNum {
		field := structVal.Field(i)
		fieldName := structType.Field(i).Name

		if field.Kind() == reflect.Pointer {
			if !field.IsNil() {
				return fmt.Errorf("field %s is not nil", fieldName)
			}
		}

		if !field.IsValid() && field.IsZero() {
			return fmt.Errorf("field %s in set", fieldName)
		}

	}

	return nil
}
