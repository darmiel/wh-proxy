package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type R struct {
	Type            string            `yaml:"type"`
	URL             string            `yaml:"url"`
	Method          string            `yaml:"method"`
	ExpectStatus    int               `yaml:"expect-status"`
	ContinueOnError bool              `yaml:"continue-on-error"`
	Data            interface{}       `yaml:"data"`
	Headers         map[string]string `yaml:"headers"`
}

type E struct {
	// Method can be either string or string-array
	Method interface{} `yaml:"method"`
	// Type can be either string or string-array
	Type      interface{} `yaml:"type"`
	Condition string      `yaml:"cond"`
}

type A struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	//
	Expect     E    `yaml:"expect"`
	BreakOnRun bool `yaml:"break-on-run"`
	//
	Data     map[string]interface{} `yaml:"data"`
	Response []R                   `yaml:"response"`
}

type Ferror string

func (e *Ferror) Error(args ...interface{}) error {
	return fmt.Errorf(string(*e), args...)
}

var (
	ErrFieldMissing Ferror = "field %s missing"
)

func ParseFile(f string) (a *A, err error) {
	var data []byte
	if data, err = os.ReadFile(f); err != nil {
		return
	}
	// unmarshal yaml
	if err = yaml.Unmarshal(data, &a); err != nil {
		return
	}
	// required fields
	if a.ID == "" {
		return nil, ErrFieldMissing.Error("id")
	}
	//
	if a.Name == "" {
		a.Name = a.ID
	}
	return
}
