package struct_viewer

// Viewer is the pkg control structure where the configuration, prefix and env vars are stored.
type Viewer struct {
	config interface{}
	prefix string

	envs []EnvVars
}

// New receives a configuration structure and a prefix and returns a Viewer struct to manipulate this library.
func New(config interface{}, prefix string) *Viewer {
	cfg := Viewer{config: config, prefix: prefix}
	cfg.Start()

	return &cfg
}

// Start starts the Viewer control struct, parsing the environment variables
func (h *Viewer) Start() {
	h.envs = parseEnvs(h.config)
}
