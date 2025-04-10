# virt-launcher
[![Release Status](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml/badge.svg)](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml)  
Launch virtual machine from cloud image using libvirt.

## dependencies
* `qemu-img`
* `virt-install`
* `virsh`

## example
```sh
# `sudo` to create file in /var/lib/libvirt/images
# `-E` to keep HOME env and cache files in $HOME/.cache
sudo -E ./virt-launcher -d \
  -n testing \
  -cpu 2 \
  -mem 4096 \
  -s 50G \
  -os debian12 \
  -key "$(cat .ssh/id_ed25519.pub)"
```
![example](./example.svg)

## cloud image
https://docs.openstack.org/image-guide/obtain-images.html
* archlinux: https://geo.mirror.pkgbuild.com/images/
* centos7.0: https://cloud.centos.org/centos/7/images/
* centos-stream9: https://cloud.centos.org/centos/9-stream/
* centos-stream10: https://cloud.centos.org/centos/10-stream/
* debian11: https://cdimage.debian.org/images/cloud/bullseye/
* debian12: https://cdimage.debian.org/images/cloud/bookworm/
* fedora41: https://download.fedoraproject.org/pub/fedora/linux/releases/41/Cloud/
* rocky8: https://dl.rockylinux.org/pub/rocky/8/images/
* rocky9: https://dl.rockylinux.org/pub/rocky/9/images/
* ubuntu18.04: https://cloud-images.ubuntu.com/bionic/
* ubuntu20.04: https://cloud-images.ubuntu.com/focal/
* ubuntu22.04: https://cloud-images.ubuntu.com/jammy/
* ubuntu24.04: https://cloud-images.ubuntu.com/noble/

## known issues
* Failed to start sshd service in archlinux:  
  https://gitlab.archlinux.org/archlinux/arch-boxes/-/issues/158
* Unknown sha256sum for latest centos7.0 aarch64 image:  
  https://cloud.centos.org/centos/7/images/sha256sum.txt
* Timeout 90s for systemd-journal-flush.service on centos-stream8 boot:  
  Failed to start Flush Journal to Persistent Storage.
