package global

type Config struct {
	Debug   bool `flag:"d,false,Enable debug output"`
	List    bool `flag:"l,false,List all cloud images"`
	Version bool `flag:"v,false,Show version and quit"`

	Distro string `flag:"n,,The distro name of cloud image"`
	Arch   string `flag:"a,,The architecture of cloud image"`
}

var CFG Config

var (
	AppName   = "virt-launcher"
	Version   = "unknown"
	BuildTime = "unknown"
)
