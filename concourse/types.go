package concourse

type Source struct {
	Target   string `json:"target"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type Version struct {
	PipelinesChecksum string `json:"pipelines_checksum"`
}

type CheckResponse []Version

type InRequest struct {
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

type InParams struct {
}

type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

type Metadata struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

type OutParams struct {
	Pipelines     []Pipeline `json:"pipelines,omitempty"`
	PipelinesFile string     `json:"pipelines_file,omitempty"`
}

type Pipeline struct {
	Name       string   `json:"name"`
	ConfigFile string   `json:"config_file"`
	VarsFiles  []string `json:"vars_files"`
}

type OutResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}
