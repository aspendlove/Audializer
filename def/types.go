package def

type Config struct {
	Folders      []string `yaml:"Folders"`
	Types        []string `yaml:"Types"`
	Workers      int      `yaml:"Workers"`
	OutputFolder string   `yaml:"OutputFolder"`
}
