module gitlab.eng.vmware.com/landerr/k8s-object-code-generator

go 1.15

require (
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.0-alpha.0
	k8s.io/apimachinery v0.22.0-alpha.0
	k8s.io/client-go v0.22.0-alpha.0
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/yaml => /Users/landerr/go/src/github.com/lander2k2/yaml
