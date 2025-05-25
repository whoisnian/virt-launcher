# virt-launcher
[![Release Status](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml/badge.svg)](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml)  
Launch virtual machine from cloud image using libvirt.

## dependencies
* `virsh`
* `virt-install`

## example
```sh
# Add your user to `libvirt` user group to access libvirt daemon.
./virt-launcher -d \
  -n testing \
  -cpu 1 \
  -mem 1024 \
  -s 20G \
  -os debian12 \
  -key "$(cat ~/.ssh/id_ed25519.pub)"
```
![example](./example.svg)

## cloud image
https://docs.openstack.org/image-guide/obtain-images.html
* Alpine Linux: https://alpinelinux.org/cloud/
* Arch Linux: https://gitlab.archlinux.org/archlinux/arch-boxes/
* CentOS: https://cloud.centos.org/centos/
* Debian: https://cdimage.debian.org/images/cloud/
* Fedora: https://fedoraproject.org/cloud/
* Rocky Linux: https://rockylinux.org/download
* Ubuntu: https://cloud-images.ubuntu.com

## known issues
* Failed to start sshd service in archlinux:  
  https://gitlab.archlinux.org/archlinux/arch-boxes/-/issues/158
* Unknown sha256sum for latest centos7.0 aarch64 image:  
  https://cloud.centos.org/centos/7/images/sha256sum.txt
* Timeout 90s for systemd-journal-flush.service on centos-stream8 boot:  
  Failed to start Flush Journal to Persistent Storage.
