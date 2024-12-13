package configs

type Repository struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
}

type Config struct {
	Repositories []Repository `json:"repositories"`
	Threads      int          `json:"threads"`
	LogFile      string       `json:"logfile"`
	WorkDir      string       `json:"workdir"`
	Debug        bool         `json:"debug"`
	Verbose      bool         `json:"verbose"`
}
