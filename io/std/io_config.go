package std

type StdIOConfig struct {
	Name  string   `json:"name"`
	Proc  string   `json:"proc"`
	Args  []string `json:"args"`
	Envs  []string `json:"envs"`
	Dir   string   `json:"dir"`
	delim string   `json:"delim"`
}
