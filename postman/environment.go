package postman

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
