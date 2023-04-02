package global

type Config struct {
	Debug   bool `flag:"d,false,Enable debug output"`
	List    bool `flag:"l,false,List all cloud images"`
	Version bool `flag:"v,false,Show version and quit"`
}

var CFG Config

var (
	Version   = "unknown"
	BuildTime = "unknown"
)
