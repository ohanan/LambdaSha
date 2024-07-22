package form

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ohanan/LambdaSha/pkg/lsha"
)

type Form struct {
	Items []*Item `json:"items"`
}
type ItemType = string

const (
	ItemTypeDesc     ItemType = "desc"
	ItemTypeCheckbox ItemType = "checkbox"
	ItemTypeRadio    ItemType = "radio"
	ItemTypeRange    ItemType = "range"
)

type TagField = string

const (
	TagFieldType            TagField = "type"
	TagFieldLabel           TagField = "label"
	TagFieldTips            TagField = "tips"
	TagFieldMax             TagField = "max"
	TagFieldMin             TagField = "min"
	TagFieldValueTipsPrefix TagField = "value_"
)

type Item struct {
	ID              string       `json:"id"`
	Type            ItemType     `json:"type"`
	Label           string       `json:"label"`
	Tips            string       `json:"tips,omitempty"`
	Checkboxes      []*CheckItem `json:"checkboxes,omitempty"`
	Radios          []*CheckItem `json:"radios,omitempty"`
	Range           *Range       `json:"range,omitempty"`
	onRadiosChecked func(label, radioLabel string)
}

type CheckItem struct {
	ID                string `json:"id"`
	Label             string `json:"label"`
	Tips              string `json:"tips,omitempty"`
	Checked           bool   `json:"checked"`
	onCheckboxChanged func(label, checkboxLabel string, checked bool)
}
type Range struct {
	Min        int64            `json:"min"`
	Max        int64            `json:"max"`
	Value      int64            `json:"value"`
	MinLabel   string           `json:"min_label"`
	MaxLabel   string           `json:"max_label"`
	ValueLabel map[int64]string `json:"value_label,omitempty"`
	onChanged  func(label string, value int)
}

var valueRegexp = regexp.MustCompile(TagFieldValueTipsPrefix + "(\\d+)")

type itemMaker interface {
	makeItem(readonly bool) *Item
}

type DescBuilder struct {
	b    *ItemsBuilder
	desc string
}

func (d *DescBuilder) Desc(desc string) lsha.ConfigDescOptionsBuilder {
	d.desc = desc
	return d
}

func (d *DescBuilder) Parent() lsha.ConfigBuilder {
	return d.b
}
func (d *DescBuilder) makeItem(readonly bool) *Item {
	if d.desc == "" {
		return nil
	}
	i := &Item{
		ID:    d.b.nextIDStr(readonly),
		Type:  ItemTypeDesc,
		Label: d.desc,
	}
	return i
}

type CheckboxBuilder struct {
	name    string
	tips    string
	b       *ItemsBuilder
	options []*CheckItem
}

func (c *CheckboxBuilder) makeItem(readonly bool) *Item {
	if c.name == "" {
		return nil
	}
	checkboxes := make([]*CheckItem, 0, len(c.options))
	for _, option := range c.options {
		label := option.Label
		if label == "" {
			continue
		}
		ci := &CheckItem{
			ID:                c.b.nextIDStr(readonly),
			Label:             label,
			Tips:              option.Tips,
			Checked:           option.Checked,
			onCheckboxChanged: option.onCheckboxChanged,
		}
		checkboxes = append(checkboxes, ci)
	}
	if len(checkboxes) == 0 {
		return nil
	}
	return &Item{
		ID:         c.b.nextIDStr(readonly),
		Type:       ItemTypeCheckbox,
		Label:      c.name,
		Tips:       c.tips,
		Checkboxes: checkboxes,
	}
}

func (c *CheckboxBuilder) SetName(name string) lsha.ConfigCheckboxOptionsBuilder {
	c.name = name
	return c
}

func (c *CheckboxBuilder) SetTips(tips string) lsha.ConfigCheckboxOptionsBuilder {
	c.tips = tips
	return c
}

func (c *CheckboxBuilder) SetOptionTips(name string, tips string) lsha.ConfigCheckboxOptionsBuilder {
	for _, option := range c.options {
		if option.Label == name {
			option.Tips = tips
		}
	}
	return c
}

func (c *CheckboxBuilder) CheckOption(name string, checked bool) lsha.ConfigCheckboxOptionsBuilder {
	for _, option := range c.options {
		if option.Label == name {
			option.Checked = checked
		}
	}
	return c
}

func (c *CheckboxBuilder) OnChangedOption(name string, onChanged func(data any, name, checkboxName string, checked bool)) lsha.ConfigCheckboxOptionsBuilder {
	for _, option := range c.options {
		if option.Label == name {
			if onChanged == nil {
				option.onCheckboxChanged = nil
			} else {
				option.onCheckboxChanged = func(label, checkboxLabel string, checked bool) {
					onChanged(c.b.data, label, checkboxLabel, checked)
				}
			}
		}
	}
	return c
}

func (c *CheckboxBuilder) ResetOptions() lsha.ConfigCheckboxOptionsBuilder {
	c.options = c.options[:0]
	return c
}

func (c *CheckboxBuilder) RemoveOption(name string) lsha.ConfigCheckboxOptionsBuilder {
	for i, option := range c.options {
		if option.Label == name {
			copy(c.options[i:], c.options[i+1:])
			c.options[len(c.options)-1] = nil
			c.options = c.options[:len(c.options)-1]
			break
		}
	}
	return c
}

func (c *CheckboxBuilder) Parent() lsha.ConfigBuilder {
	return c.b
}

func (c *CheckboxBuilder) AddOption(name string, tips string, checked bool, onChanged func(data any, name, checkboxName string, checked bool)) lsha.ConfigCheckboxOptionsBuilder {
	i := &CheckItem{
		Label:   name,
		Tips:    tips,
		Checked: checked,
	}
	if onChanged != nil {
		i.onCheckboxChanged = func(label, checkboxLabel string, checked bool) {
			onChanged(c.b.data, label, checkboxLabel, checked)
		}
	}
	c.options = append(c.options, i)
	return c
}

type RadioBuilder struct {
	name      string
	tips      string
	options   []*CheckItem
	b         *ItemsBuilder
	onChecked func(name, radioName string)
}

func (r *RadioBuilder) makeItem(readonly bool) *Item {
	if r.name == "" {
		return nil
	}
	radios := make([]*CheckItem, 0, len(r.options))
	var hasChecked bool
	for _, option := range r.options {
		if option.Label == "" {
			continue
		}
		thisChecked := !hasChecked && option.Checked
		if thisChecked {
			hasChecked = true
		}
		radios = append(radios, &CheckItem{
			ID:      r.b.nextIDStr(readonly),
			Label:   option.Label,
			Tips:    option.Tips,
			Checked: thisChecked,
		})
	}
	if len(radios) == 0 {
		return nil
	}
	onChecked := r.onChecked
	return &Item{
		ID:     r.b.nextIDStr(readonly),
		Type:   ItemTypeRadio,
		Label:  r.name,
		Tips:   r.tips,
		Radios: radios,
		onRadiosChecked: func(label, radioLabel string) {
			onChecked(label, radioLabel)
		},
	}
}

func (r *RadioBuilder) Parent() lsha.ConfigBuilder {
	return r.b
}

func (r *RadioBuilder) SetName(name string) lsha.ConfigRadioOptionsBuilder {
	r.name = name
	return r
}

func (r *RadioBuilder) SetTips(tips string) lsha.ConfigRadioOptionsBuilder {
	r.tips = tips
	return r
}

func (r *RadioBuilder) AddOption(name string, tips string) lsha.ConfigRadioOptionsBuilder {
	r.options = append(r.options, &CheckItem{
		Label: name,
		Tips:  tips,
	})
	return r
}

func (r *RadioBuilder) CheckOption(name string) lsha.ConfigRadioOptionsBuilder {
	for _, option := range r.options {
		option.Checked = option.Label == name
	}
	return r
}

func (r *RadioBuilder) ResetOptions() lsha.ConfigRadioOptionsBuilder {
	r.options = r.options[:0]
	return r
}

func (r *RadioBuilder) RemoveOption(name string) lsha.ConfigRadioOptionsBuilder {
	for i, option := range r.options {
		if option.Label == name {
			copy(r.options[i:], r.options[i+1:])
			r.options[len(r.options)-1] = nil
			r.options = r.options[:len(r.options)-1]
			break
		}
	}
	return r
}

func (r *RadioBuilder) OnCheckedOption(onChecked func(data any, name, radioName string)) lsha.ConfigRadioOptionsBuilder {
	if onChecked == nil {
		r.onChecked = nil
	} else {
		r.onChecked = func(name, radioName string) {
			onChecked(r.b.data, name, radioName)
		}
	}
	return r
}

type RangeBuilder struct {
	b    *ItemsBuilder
	r    *Range
	name string
	tips string
}

func (r *RangeBuilder) makeItem(readonly bool) *Item {
	if r.name == "" {
		return nil
	}
	copiedValueTips := map[int64]string{}
	for i, s := range r.r.ValueLabel {
		copiedValueTips[i] = s
	}
	return &Item{
		ID:    r.b.nextIDStr(readonly),
		Type:  ItemTypeRange,
		Label: r.name,
		Tips:  r.tips,
		Range: &Range{
			Min:        r.r.Min,
			Max:        r.r.Max,
			Value:      r.r.Value,
			MinLabel:   r.r.MinLabel,
			MaxLabel:   r.r.MaxLabel,
			ValueLabel: copiedValueTips,
			onChanged:  r.r.onChanged,
		},
	}
}

func (r *RangeBuilder) OnChanged(onChanged func(data any, name string, value int)) lsha.ConfigRangeOptionsBuilder {
	if onChanged == nil {
		r.r.onChanged = nil
	} else {
		r.r.onChanged = func(label string, value int) {
			onChanged(r.b.data, label, value)
		}
	}
	return r
}

func (r *RangeBuilder) RemoveValueTip(value int) lsha.ConfigRangeOptionsBuilder {
	delete(r.r.ValueLabel, int64(value))
	return r
}

func (r *RangeBuilder) Parent() lsha.ConfigBuilder {
	return r.b
}

func (r *RangeBuilder) SetName(name string) lsha.ConfigRangeOptionsBuilder {
	r.name = name
	return r
}

func (r *RangeBuilder) SetTips(tips string) lsha.ConfigRangeOptionsBuilder {
	r.tips = tips
	return r
}

func (r *RangeBuilder) Max(value int, label string) lsha.ConfigRangeOptionsBuilder {
	r.r.Max, r.r.MaxLabel = int64(value), label
	return r
}

func (r *RangeBuilder) Min(value int, label string) lsha.ConfigRangeOptionsBuilder {
	r.r.Min, r.r.MinLabel = int64(value), label
	return r
}

func (r *RangeBuilder) ValueTips(value int, tips string) lsha.ConfigRangeOptionsBuilder {
	r.r.ValueLabel[int64(value)] = tips
	return r
}

func (r *RangeBuilder) Value(value int) lsha.ConfigRangeOptionsBuilder {
	r.r.Value = int64(value)
	return r
}

type ItemsBuilder struct {
	data       any
	itemMakers []itemMaker
	nextID     int
}

func (b *ItemsBuilder) BindData(data any) {
	b.data = data
}
func (b *ItemsBuilder) Data() any {
	return b.data
}

func (b *ItemsBuilder) Checkbox(name string, tips string) lsha.ConfigCheckboxOptionsBuilder {
	bb := &CheckboxBuilder{
		name: name,
		tips: tips,
		b:    b,
	}
	return bb
}

func (b *ItemsBuilder) Radio(name string, tips string) lsha.ConfigRadioOptionsBuilder {
	bb := &RadioBuilder{
		name: name,
		tips: tips,
		b:    b,
	}
	return bb
}

func (b *ItemsBuilder) Range(name string, tips string) lsha.ConfigRangeOptionsBuilder {
	return &RangeBuilder{
		b: b,
		r: &Range{
			ValueLabel: map[int64]string{},
		},
		name: name,
		tips: tips,
	}
}

func (b *ItemsBuilder) nextIDStr(readonly bool) string {
	if readonly {
		return ""
	}
	b.nextID++
	return strconv.Itoa(b.nextID)
}
func (b *ItemsBuilder) Desc(desc string) lsha.ConfigDescOptionsBuilder {
	bb := &DescBuilder{
		b:    b,
		desc: desc,
	}
	b.itemMakers = append(b.itemMakers, bb)
	return bb
}
func (b *ItemsBuilder) Build(readonly bool) []*Item {
	items := make([]*Item, 0, len(b.itemMakers))
	for _, maker := range b.itemMakers {
		if i := maker.makeItem(readonly); i != nil {
			items = append(items, i)
		}
	}
	return items
}

func ConvertItems(v any) (items []*Item, err error) {
	rt := reflect.TypeOf(v)
	handler, err := getHandler(rt)
	if err != nil {
		return nil, err
	}
	items = make([]*Item, len(handler.itemMakers))
	rv := reflect.ValueOf(v).Elem()
	for i, maker := range handler.itemMakers {
		items[i] = maker(rv)
	}
	return
}
func UpdateItems(items []*Item, data map[string]any) {
	for _, item := range items {
		switch item.Type {
		case ItemTypeDesc:
		case ItemTypeCheckbox:
			for _, checkbox := range item.Checkboxes {
				v, ok := data[item.ID+"."+checkbox.ID]
				if ok {
					if v, ok := v.(bool); ok {
						old := checkbox.Checked
						checkbox.Checked = v
						if old != v && checkbox.onCheckboxChanged != nil {
							checkbox.onCheckboxChanged(item.Label, checkbox.Label, v)
						}
					}
				}
			}
		case ItemTypeRadio:
			for _, radio := range item.Radios {
				v, ok := data[item.ID+"."+radio.ID]
				if ok {
					if v, ok := v.(bool); ok && v {
						var notifyCheckedLabel string
						for _, radio2 := range item.Radios {
							oldChecked := radio2.Checked
							radio2.Checked = radio2.ID == radio.ID
							if !oldChecked && radio2.Checked {
								notifyCheckedLabel = radio2.Label
							}
						}
						if notifyCheckedLabel != "" && item.onRadiosChecked != nil {
							item.onRadiosChecked(item.Label, notifyCheckedLabel)
						}
						break
					}
				}
			}
		case ItemTypeRange:
			i, err := strconv.ParseInt(fmt.Sprint(data[item.ID]), 10, 64)
			if err == nil {
				r := item.Range
				old := r.Value
				if r.Min <= i && i <= r.Max {
					r.Value = i
					if old != i && r.onChanged != nil {
						r.onChanged(item.Label, int(i))
					}
				}
			}
		}
	}
}
func UpdateForm(v any, data map[string]any) (err error) {
	rt := reflect.TypeOf(v)
	handler, err := getHandler(rt)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	for k, v := range data {
		if setter, ok := handler.valueSetters[k]; ok {
			err = setter(rv, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Handler struct {
	itemMakers   []func(v reflect.Value) *Item
	error        error
	valueSetters map[string]func(v reflect.Value, data any) error
}

var handlerCache sync.Map

func getHandler(rt reflect.Type) (*Handler, error) {
	c, ok := handlerCache.Load(rt)
	if ok {
		h := c.(*Handler)
		if h.error != nil {
			return nil, h.error
		}
		return h, nil
	}
	h := &Handler{
		valueSetters: map[string]func(v reflect.Value, data any) error{},
	}
	defer func() {
		handlerCache.Store(rt, h)
	}()
	if rt.Kind() != reflect.Pointer || rt.Elem().Kind() != reflect.Struct {
		h.error = errors.New("form must be a pointer to a struct")
		return h, nil
	}
	rt = rt.Elem()
	nf := rt.NumField()
	for i := 0; i < nf; i++ {
		f := rt.Field(i)
		if !f.IsExported() || f.Anonymous {
			continue
		}
		typeTag, ok := f.Tag.Lookup(TagFieldType)
		if !ok {
			continue
		}
		name := f.Name
		label := f.Tag.Get(TagFieldLabel)
		tips := f.Tag.Get(TagFieldTips)
		var checkItems []*CheckItem
		if f.Type.Kind() == reflect.Struct {
			snf := f.Type.NumField()
			for i := 0; i < snf; i++ {
				field := f.Type.Field(i)
				if !field.IsExported() || field.Anonymous || field.Type.Kind() != reflect.Bool {
					continue
				}
				fieldLabel, ok := field.Tag.Lookup(TagFieldLabel)
				if !ok {
					continue
				}
				checkItems = append(checkItems, &CheckItem{
					ID:    field.Name,
					Label: fieldLabel,
					Tips:  field.Tag.Get(TagFieldTips),
				})
			}
		}
		switch typeTag {
		case ItemTypeDesc:
			if f.Type.Kind() != reflect.String {
				continue
			}
			h.itemMakers = append(h.itemMakers, func(v reflect.Value) *Item {
				return &Item{
					ID:    name,
					Type:  ItemTypeDesc,
					Label: v.FieldByName(name).String(),
				}
			})
		case ItemTypeCheckbox:
			for _, item := range checkItems {
				id := name + "." + item.ID
				h.valueSetters[id] = func(v reflect.Value, data any) error {
					if b, ok := data.(bool); ok {
						v.Elem().FieldByName(name).FieldByName(item.ID).SetBool(b)
					} else {
						return fmt.Errorf("expect bool value, but got %T", data)
					}
					return nil
				}
			}
			if len(checkItems) > 0 {
				h.itemMakers = append(h.itemMakers, func(v reflect.Value) *Item {
					checkboxes := make([]*CheckItem, len(checkItems))
					v = v.FieldByName(name)
					for idx, item := range checkItems {
						checkboxes[idx] = &CheckItem{
							ID:      name + "." + item.ID,
							Label:   item.Label,
							Tips:    item.Tips,
							Checked: v.FieldByName(item.ID).Bool(),
						}
					}
					return &Item{
						ID:         name,
						Type:       ItemTypeCheckbox,
						Label:      label,
						Tips:       tips,
						Checkboxes: checkboxes,
					}
				})
			}
		case ItemTypeRadio:
			for _, item := range checkItems {
				item := item
				id := name + "." + item.ID
				h.valueSetters[id] = func(v reflect.Value, data any) error {
					if b, ok := data.(bool); ok && b {
						for _, checkItem := range checkItems {
							if checkItem.ID != item.ID {
								v.Elem().FieldByName(name).FieldByName(checkItem.ID).SetBool(false)
							} else {
								v.Elem().FieldByName(name).FieldByName(checkItem.ID).SetBool(true)
							}
						}
					} else {
						return fmt.Errorf("expect true, but got %v", data)
					}
					return nil
				}
			}
			if len(checkItems) > 0 {
				h.itemMakers = append(h.itemMakers, func(v reflect.Value) *Item {
					radios := make([]*CheckItem, len(checkItems))
					v = v.FieldByName(name)
					for idx, item := range checkItems {
						radios[idx] = &CheckItem{
							ID:      name + "." + item.ID,
							Label:   item.Label,
							Tips:    item.Tips,
							Checked: v.FieldByName(item.ID).Bool(),
						}
					}
					return &Item{
						ID:     name,
						Type:   ItemTypeRadio,
						Label:  label,
						Tips:   tips,
						Radios: radios,
					}
				})
			}
		case ItemTypeRange:
			if f.Type.Kind() != reflect.Int {
				continue
			}
			minValue, minLabel := findRangeLimit(f, TagFieldMin)
			maxValue, maxLabel := findRangeLimit(f, TagFieldMax)
			valueLabels := map[int64]string{}
			for _, s := range valueRegexp.FindAllStringSubmatch(string(f.Tag), -1) {
				valueInt, err := strconv.ParseInt(s[1], 10, 64)
				if err == nil {
					valueLabels[valueInt] = f.Tag.Get(fmt.Sprintf("%s%d", TagFieldValueTipsPrefix, valueInt))
				}
			}
			h.valueSetters[name] = func(v reflect.Value, data any) error {
				vv, err := strconv.ParseInt(fmt.Sprint(data), 10, 64)
				if err != nil {
					return fmt.Errorf("expect int value, but got %T", data)
				}
				if maxValue > 0 && vv > maxValue {
					return fmt.Errorf("value(%d) is too large, expect less than %d", vv, maxValue)
				}
				if minValue > 0 && vv < minValue {
					return fmt.Errorf("value(%d) is too small, expect greater than %d", vv, minValue)
				}
				v.Elem().FieldByName(name).SetInt(vv)
				return nil
			}
			h.itemMakers = append(h.itemMakers, func(v reflect.Value) *Item {
				return &Item{
					ID:    name,
					Type:  ItemTypeRange,
					Label: label,
					Tips:  tips,
					Range: &Range{
						Min:        minValue,
						Max:        maxValue,
						Value:      v.FieldByName(name).Int(),
						MinLabel:   minLabel,
						MaxLabel:   maxLabel,
						ValueLabel: valueLabels,
					},
				}
			})
		}
	}
	return h, nil
}
func findRangeLimit(tf reflect.StructField, name string) (value int64, label string) {
	if v, ok := tf.Tag.Lookup(name); ok {
		var valueStr string
		valueStr, label = split2(v)
		value, _ = strconv.ParseInt(valueStr, 10, 64)
	}
	return
}

func split2(v string) (v1, v2 string) {
	split := strings.SplitN(v, ",", 2)
	if len(split) > 0 {
		v1 = strings.TrimSpace(split[0])
		if len(split) > 1 {
			v2 = strings.TrimSpace(split[1])
		}
	}
	return
}
