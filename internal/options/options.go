package options

type GenerateOption int

const (
	WithMarkers GenerateOption = iota
	WithYAML
	WithGo
)

type RBACOptions struct {
	ManifestFilepaths []string
	ManifestFilepath  string
	RoleName          string
	VariableName      string
	ValuesFilePath    string
	Verbs             []string
	UseResourceNames  bool
}
