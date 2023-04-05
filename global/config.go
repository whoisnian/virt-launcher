package global

type Config struct {
	Debug bool `flag:"d,false,Enable debug output"`

	ListAll bool `flag:"l,false,List all cloud images"`
	DryRun  bool `flag:"dry-run,false,Prepare resources without creating vm"`
	Version bool `flag:"v,false,Show version and quit"`

	Os   string `flag:"os,,The distribution name of cloud image"`
	Arch string `flag:"arch,,The architecture of cloud image"`
}

var CFG Config

var (
	AppName   = "virt-launcher"
	Version   = "unknown"
	BuildTime = "unknown"
)
