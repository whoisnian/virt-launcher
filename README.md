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
* centos-stream8: https://cloud.centos.org/centos/8-stream/x86_64/images/
* centos-stream9: https://cloud.centos.org/centos/9-stream/x86_64/images/
* debian10: https://cdimage.debian.org/images/cloud/buster/
* debian11: https://cdimage.debian.org/images/cloud/bullseye/
* debian12: https://cdimage.debian.org/images/cloud/bookworm/
* ubuntu18.04: https://cloud-images.ubuntu.com/bionic/
* ubuntu20.04: https://cloud-images.ubuntu.com/focal/
* ubuntu22.04: https://cloud-images.ubuntu.com/jammy

## known issues
* Failed to start sshd service in archlinux:  
  https://gitlab.archlinux.org/archlinux/arch-boxes/-/issues/158
* Timeout 90s for systemd-journal-flush.service on centos-stream8 boot:  
  Failed to start Flush Journal to Persistent Storage.
