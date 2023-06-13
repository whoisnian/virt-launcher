# virt-launcher
[![Release Status](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml/badge.svg)](https://github.com/whoisnian/virt-launcher/actions/workflows/release.yml)  
Launch virtual machine from cloud image using libvirt.

## dependencies
* `qemu-img`
* `virt-install`
* `virsh`

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
* Unknown sha256sum for latest centos7.0 image:  
  https://cloud.centos.org/centos/7/images/sha256sum.txt
* Timeout 90s for systemd-journal-flush.service on centos-stream8 boot:  
  Failed to start Flush Journal to Persistent Storage.
