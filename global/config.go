package global

import (
	"context"

	"github.com/whoisnian/glb/config"
)

var CFG Config

type Config struct {
	Debug bool `flag:"d,false,Enable debug output"`

	Update  bool `flag:"u,false,Update the index of cloud images and quit"`
	ListAll bool `flag:"l,false,List all supported cloud images and quit"`
	Prepare bool `flag:"pre,false,Prepare resources without creating vm"`
	Version bool `flag:"v,false,Show version and quit"`

	Name string `flag:"n,testing,Name of the guest vm"`
	Os   string `flag:"os,,The distribution name of cloud image"`
	Arch string `flag:"arch,,The architecture of cloud image"`
	Size string `flag:"s,20G,Resize guest vm disk image to"`
	Boot string `flag:"boot,hd,Boot options for the guest vm"`
	Cpu  string `flag:"cpu,1,Number of vCPUs for the guest vm"`
	Mem  string `flag:"mem,1024,Memory allocated to the guest vm"`
	Key  string `flag:"key,,Authorized keys for default user"`
	Pass string `flag:"pass,,Login password for default user"`

	Network string `flag:"network,default,Network name for the guest vm"`
	Storage string `flag:"storage,default,Storage pool name for the disk image"`
	Connect string `flag:"connect,qemu:///system,Connect to hypervisor with libvirt URI"`
}

func SetupConfig(_ context.Context) {
	_, err := config.FromCommandLine(&CFG)
	if err != nil {
		panic(err)
	}
}
