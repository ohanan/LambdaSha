package form

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Form struct {
	Items []*Item `json:"items"`
}
type Item struct {
	Content    string       `json:"content"`
	Label      string       `json:"label"`
	Tips       string       `json:"tips,omitempty"`
	Checkboxes []*CheckItem `json:"checkboxes,omitempty"`
	Radios     []*CheckItem `json:"radios,omitempty"`
	Range      *Range       `json:"range,omitempty"`
}

type CheckItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Tips    string `json:"tips,omitempty"`
	Checked bool   `json:"checked"`
}
type Range struct {
	Min       int            `json:"min"`
	Max       int            `json:"max"`
	Value     int            `json:"value"`
	MinLabel  string         `json:"min_label"`
	MaxLabel  string         `json:"max_label"`
	ValueTips map[int]string `json:"value_tips,omitempty"`
}

type Test struct {
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
	Range int `type:"range" label:"Range" mix:"5,min value" max:"10,max value" value_6:"value of 6" value_7:"value of 7"`
}

var valueRegexp = regexp.MustCompile("value_(\\d+)")

func ConvertItems(v any) ([]*Item, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("form must be a struct")
	}
	nf := rv.NumField()
	rt := rv.Type()
	var items []*Item
	for i := 0; i < nf; i++ {
		tf := rt.Field(i)
		if !tf.IsExported() {
			continue
		}
		if tf.Anonymous {
			return nil, errors.New("form should not have anonymous fields")
		}
		typ, ok := tf.Tag.Lookup("type")
		if !ok {
			continue
		}
		vf := rv.Field(i)
		item := &Item{
			Content:    "",
			Label:      "",
			Tips:       "",
			Checkboxes: nil,
			Radios:     nil,
		}
		switch tf.Type.Kind() {
		case reflect.Struct:
			switch typ {
			case "checkbox":
			case "radio":
			}
		case reflect.Int:
			if typ == "range" {
				r := &Range{
					Value: int(vf.Int()),
				}
				item.Range = r
				r.Min, r.MinLabel = findRangeMinOrMax(tf, "min")
				r.Max, r.MaxLabel = findRangeMinOrMax(tf, "max")
				for _, ss := range valueRegexp.FindAllStringSubmatch(string(tf.Tag), -1) {
					if len(ss) < 2 {
						continue
					}
					value, _ := strconv.Atoi(ss[1])
					if value < item.Range.Min || value > item.Range.Max {
						continue
					}
					item.Range.ValueTips[value] = strings.TrimSpace(tf.Tag.Get(ss[0]))
				}
				items = append(items, item)
			}
		case reflect.String:
		default:

		}

	}
}
func findCheckItem(tf reflect.StructField, name string)
func findRangeMinOrMax(tf reflect.StructField, name string) (value int, label string) {
	if v, ok := tf.Tag.Lookup(name); ok {
		split := strings.Split(v, ",")
		if len(split) > 0 {
			value, _ = strconv.Atoi(strings.TrimSpace(split[0]))
			if len(split) > 1 {
				label = strings.TrimSpace(split[1])
			}
		}
	}
	return
}
