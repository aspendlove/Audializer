package def

type Config struct {
	Folders      []string `yaml:"Folders"`
	Workers      int      `yaml:"Workers"`
	OutputFolder string   `yaml:"OutputFolder"`
}
