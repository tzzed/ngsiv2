package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Context entity: a thing in the NGSI model.
type Entity struct {
	Id         string                `json:"id"`
	Type       string                `json:"type,omitempty"`
	Attributes map[string]*Attribute `json:"-"`
}

type typeValue struct {
	//Name  string      `json:"name"`
	Type  AttributeType `json:"type,omitempty"`
	Value interface{}   `json:"value"`
}

// Context attribute: property of a context entity.
type Attribute struct {
	typeValue
	Metadata map[string]*Metadata `json:"metadata"`
}

// Context metadata: an optional part of the attribute.
type Metadata struct {
	typeValue
}

type AttributeType string

const (
	StringType     AttributeType = "String"
	FloatType      AttributeType = "Float"
	IntegerType    AttributeType = "Integer"
	PercentageType AttributeType = "Percentage"
	DateTimeType   AttributeType = "DateTime"
)

type ActionType string

const (
	AppendAction       ActionType = "append"
	AppendStrictAction ActionType = "appendStrict"
	UpdateAction       ActionType = "update"
	DeleteAction       ActionType = "delete"
	ReplaceAction      ActionType = "replace"
)

type BatchUpdate struct {
	ActionType ActionType `json:"actionType"`
	Entities   []*Entity  `json:"entities"`
}

// Creates a new context entity with id and type and no attributes.
func NewEntity(id string, entityType string) *Entity {
	e := &Entity{}
	e.Id = id
	e.Type = entityType
	e.Attributes = make(map[string]*Attribute)
	return e
}

type _entity Entity

func (e *Entity) UnmarshalJSON(b []byte) error {
	t_ := _entity{}
	if err := json.Unmarshal(b, &t_); err != nil {
		return err
	}

	_ = json.Unmarshal(b, &(t_.Attributes))
	/*if err := json.Unmarshal(b, &(t_.Attributes)); err != nil {
		return err
	}*/

	typ := reflect.TypeOf(t_)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			delete(t_.Attributes, jsonTag)
		}
	}

	*e = Entity(t_)

	return nil
}

func (e *Entity) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	for k, v := range e.Attributes {
		data[k] = v
	}

	// Take all the struct values with a json tag
	val := reflect.ValueOf(*e)
	typ := reflect.TypeOf(*e)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldv := val.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			data[jsonTag] = fieldv.Interface()
		}
	}

	return json.Marshal(data)
}

func (e *Entity) GetAttribute(name string) (*Attribute, error) {
	if attr, ok := e.Attributes[name]; ok {
		return attr, nil
	} else {
		return nil, fmt.Errorf("Entity has no attribute '%s'", name)
	}
}

func (e *Entity) SetAttributeAsString(name string, value string) {
	e.Attributes[name] = &Attribute{
		typeValue: typeValue{
			Type:  StringType,
			Value: value,
		},
	}
}

func (e *Entity) SetAttributeAsInteger(name string, value int) {
	e.Attributes[name] = &Attribute{
		typeValue: typeValue{
			Type:  IntegerType,
			Value: value,
		},
	}
}

func (e *Entity) SetAttributeAsFloat(name string, value float64) {
	e.Attributes[name] = &Attribute{
		typeValue: typeValue{
			Type:  FloatType,
			Value: value,
		},
	}
}

func (a *Attribute) GetAsString() (string, error) {
	if a.Type != StringType {
		return "", fmt.Errorf("Attribute is not String, but %s", a.Type)
	}
	return a.Value.(string), nil
}

func (a *Attribute) GetAsInteger() (int, error) {
	if a.Type != IntegerType {
		return 0, fmt.Errorf("Attribute is not Integer, but %s", a.Type)
	}
	return int(a.Value.(float64)), nil
}

func (a *Attribute) GetAsFloat() (float64, error) {
	if a.Type != FloatType {
		return 0, fmt.Errorf("Attribute is not Float, but %s", a.Type)
	}
	return a.Value.(float64), nil
}

func NewBatchUpdate(action ActionType) *BatchUpdate {
	b := &BatchUpdate{ActionType: action}
	return b
}

func (u *BatchUpdate) AddEntity(entity *Entity) {
	u.Entities = append(u.Entities, entity)
}
