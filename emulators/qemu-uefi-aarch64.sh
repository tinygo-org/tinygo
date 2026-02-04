#!/bin/sh

set -e

binary=$1
dir=$(dirname $binary)

mkdir -p ${dir}/qemu-disk/efi/boot
mv $binary ${dir}/qemu-disk/efi/boot/bootaa64.efi

ovmf_path=${OVMF_CODE:-/usr/share/AAVMF/AAVMF_CODE.no-secboot.fd}

exec qemu-system-aarch64 -M virt -bios $ovmf_path -device virtio-rng-pci \
	-cpu max,pauth-impdef=on -drive format=raw,file=fat:rw:${dir}/qemu-disk \
	-net none -nographic -no-reboot
