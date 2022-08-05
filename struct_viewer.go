package struct_viewer

type Viewer struct {
	config interface{}
	prefix string

	envs []EnvVars
}

func New(config interface{}, prefix string) *Viewer {
	cfg := Viewer{config: config, prefix: prefix}
	cfg.Start()

	return &cfg
}

func (h *Viewer) Start() {
	h.envs = parseEnvs(h.config)
}
