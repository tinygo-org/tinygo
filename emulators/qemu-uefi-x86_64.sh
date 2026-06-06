#!/bin/sh

set -e

binary=$1
dir=$(dirname $binary)

mkdir -p ${dir}/qemu-disk/efi/boot
mv $binary ${dir}/qemu-disk/efi/boot/bootx64.efi

ovmf_path=${OVMF_CODE:-/usr/share/qemu/OVMF.fd}

exec qemu-system-x86_64 -bios $ovmf_path -cpu qemu64,+rdrand -device virtio-rng-pci \
	-drive format=raw,file=fat:rw:${dir}/qemu-disk -net none -nographic -no-reboot
