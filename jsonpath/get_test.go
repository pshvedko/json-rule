package jsonpath_test

import (
	"encoding/json"
	"fmt"
	"github.com/pshvedko/json-rule/jsonpath"
	"reflect"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	var j = map[string]interface{}{
		"access_level": "high",
		"id": map[string]interface{}{
			"value": 15,
		},
		"num": map[string]interface{}{
			"value": map[string]interface{}{
				"all":     15,
				"smaller": 10,
				"status":  "ok",
			},
		},
		"object_array": map[string]interface{}{
			"field": "value",
			"list": []interface{}{
				map[string]interface{}{
					"all":    "good",
					"id":     1,
					"status": "ok",
				},
				map[string]interface{}{
					"all":    "better",
					"id":     2,
					"status": "ok",
				},
				map[string]interface{}{
					"all":    "best",
					"id":     3,
					"status": "ok",
				},
			},
		},
		"statuses": []interface{}{
			map[string]interface{}{
				"all":       "good",
				"id":        1,
				"status_id": 4,
			},
			map[string]interface{}{
				"all":       "better",
				"id":        2,
				"status_id": 3,
			},
			map[string]interface{}{
				"all":       "best",
				"id":        3,
				"status_id": 1,
			},
		},
	}
	type args struct {
		j interface{}
		k []string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "object_array.list.#.status",
			args: args{
				j: j,
				k: []string{"object_array", "list", "#", "status"},
			},
			want:    []interface{}{"ok", "ok", "ok"},
			wantErr: false,
		}, {
			name: "statuses.#.status_id",
			args: args{
				j: j,
				k: []string{"statuses", "#", "status_id"},
			},
			want:    []interface{}{4, 3, 1},
			wantErr: false,
		}, {
			name: "num.value.all",
			args: args{
				j: j,
				k: []string{"num", "value", "all"},
			},
			want:    15,
			wantErr: false,
		}, {
			name: "id.value",
			args: args{
				j: j,
				k: []string{"id", "value"},
			},
			want:    15,
			wantErr: false,
		}, {
			name: "access_level",
			args: args{
				j: j,
				k: []string{"access_level"},
			},
			want:    "high",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonpath.Get(tt.args.j, tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleGet() {
	var j, v interface{}
	err := json.NewDecoder(strings.NewReader(
		`{ "a": { "b": [ { "c": 1 }, { "c": 2 }, { "c": 3, "x": true }, { "x": true } ] } }`)).Decode(&j)
	if err != nil {
		return
	}
	v, err = jsonpath.Get(j, []string{"a", "b", "#", "c"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(v)
	// Output:
	// [1 2 3]
}
