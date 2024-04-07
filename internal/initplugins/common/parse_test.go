package common

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type obj struct {
		Field1       string `required:"true"`
		Field2       string `required:"false"`
		Field3       string
		Field4String string `json:"field4"`
	}

	type args struct {
		obj    *obj
		params map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    obj
		wantErr bool
	}{
		{
			name: "all fields are provided",
			args: args{
				obj: &obj{},
				params: map[string]string{
					"Field1": "value1",
					"Field2": "value2",
					"Field3": "value3",
				},
			},
			want: obj{
				Field1: "value1",
				Field2: "value2",
				Field3: "value3",
			},
			wantErr: false,
		},
		{
			name: "required field is missing",
			args: args{
				obj: &obj{},
				params: map[string]string{
					"Field2": "value2",
					"Field3": "value3",
				},
			},
			wantErr: true,
		},
		{
			name: "optional field is missing",
			args: args{
				obj: &obj{},
				params: map[string]string{
					"Field1": "value1",
				},
			},
			want: obj{
				Field1: "value1",
				Field2: "",
				Field3: "",
			},
			wantErr: false,
		},
		{
			name: "json tag",
			args: args{
				obj: &obj{},
				params: map[string]string{
					"Field1": "value1",
					"Field2": "value2",
					"Field3": "value3",
					"field4": "value4",
				},
			},
			want: obj{
				Field1:       "value1",
				Field2:       "value2",
				Field3:       "value3",
				Field4String: "value4",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Parse(tt.args.obj, tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && reflect.DeepEqual(tt.args.obj, tt.want) {
				t.Errorf("Parse() = %v, want %v", tt.args.obj, tt.want)
			}

			fmt.Printf("%#v\n", tt.args.obj)
		})
	}
}
