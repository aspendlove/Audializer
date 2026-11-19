package def

type Config struct {
	Folders []FolderPair `yaml:"Folders"`
	Workers int          `yaml:"Workers"`
}

type FolderPair struct {
	Source      string `yaml:"Source"`
	Destination string `yaml:"Destination"`
}
