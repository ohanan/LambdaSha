package form

import (
	"encoding/json"
	"testing"
)

type TestData struct {
	Description string `type:"desc"`
	Checkboxes  struct {
		Select1 bool `label:"select1" tips:"tips for select1"`
		Select2 bool `label:"select2" tips:"tips for select2"`
		Select3 bool `label:"select3" tips:"tips for select3"`
	} `type:"checkbox" label:"Checkboxes" tips:"tips for Checkboxes"`
	Radios struct {
		Select1 bool `label:"select1" tips:"tips for select1"`
		Select2 bool `label:"select2" tips:"tips for select2"`
		Select3 bool `label:"select3" tips:"tips for select3"`
	} `type:"radio" label:"Radios" tips:"tips for Radios"`
	Range int `type:"range" label:"Range" min:"5,min value" max:"10,max value" value_6:"value of 6" value_7:"value of 7"`
}

func TestConvertItems(t *testing.T) {
	td := &TestData{}
	td.Description = "this is a description"
	td.Checkboxes.Select3 = true
	td.Radios.Select3 = true
	td.Range = 6
	items, err := ConvertItems(td)
	if err != nil {
		t.Fatal(err)
	}
	p := func(v any) {
		x, _ := json.MarshalIndent(v, "", "  ")
		t.Log(string(x))
	}
	p(items)
	UpdateForm(td, map[string]any{
		// "Range":              10,
		// "Checkboxes.Select2": true,
		// "Checkboxes.Select3": false,
		"Radios.Select2": true,
	})
	p(td)
}
