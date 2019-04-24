package postman

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
)

// Environment defines an execution environment
type Environment struct {
	EnvMeta
	ID     string
	Name   string
	Values []Variable
}

// Variable defines a defined variable value
type Variable struct {
	Key         string
	Value       string
	Type        string
	Description string
	Enabled     bool
}

// EnvMeta defines the metadata for an environment
type EnvMeta struct {
	VariableScope string `json:"_postman_variable_scope"`
	ExportedAt    string `json:"_postman_exported_at"`
	ExportedUsing string `json:"_postman_exported_using"`
}

// EnvironmentFromFile creates an environment from a file
func EnvironmentFromFile(filepath string) (*Environment, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	env := Environment{}
	if err := json.Unmarshal(file, &env); err != nil {
		return nil, err
	}

	return &env, nil
}

// VariableMap returns a msp[string]string of the environment's variables
func (e *Environment) VariableMap() map[string]string {
	varMap := make(map[string]string)

	for _, v := range e.Values {
		if v.Enabled {
			varMap[v.Key] = v.Value
		}
	}

	return varMap
}

// SubstVars subsitutes variables in a string
func SubstVars(templ string, vars map[string]string) (string, error) {
	tmpl, err := template.New("gopherman").Parse(templ)
	if err != nil {
		return "", err
	}

	output := bytes.NewBuffer([]byte{})

	err = tmpl.Execute(output, vars)
	if err != nil {
		panic(err)
	}

	outBytes, err := ioutil.ReadAll(output)
	if err != nil {
		return "", err
	}

	return string(outBytes), nil
}
