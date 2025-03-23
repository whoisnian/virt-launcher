package global

import (
	"context"

	"github.com/whoisnian/glb/config"
)

var CFG Config

type Config struct {
	Debug bool `flag:"d,false,Enable debug output"`

	ListAll bool `flag:"l,false,List all supported cloud images"`
	DryRun  bool `flag:"dry-run,false,Prepare resources without creating vm"`
	Version bool `flag:"v,false,Show version and quit"`

	Name string `flag:"n,testing,Name of the guest vm"`
	Os   string `flag:"os,,The distribution name of cloud image"`
	Arch string `flag:"arch,,The architecture of cloud image"`
	Size string `flag:"s,20G,Resize vm disk image to"`
	Cpu  string `flag:"cpu,1,Number of vCPUs for the guest vm"`
	Mem  string `flag:"mem,1024,Memory allocated to the guest vm"`
	Key  string `flag:"key,,Authorized keys for default user"`
	Pass string `flag:"pass,,Login password for default user"`

	Storage string `flag:"storage,/var/lib/libvirt/images,Directory of libvirt storage pool"`
	Connect string `flag:"connect,qemu:///system,Connect to hypervisor with libvirt URI"`
}

func SetupConfig(_ context.Context) {
	_, err := config.FromCommandLine(&CFG)
	if err != nil {
		panic(err)
	}
}
