set -e
if [ ! -f '/Volumes/BOOTLOADER/DETAILS.TXT' ]; then
    echo "MAX32620 not mounted. Mounting..."
    [ -d  /Volumes/BOOTLOADER ] || sudo mkdir /Volumes/BOOTLOADER
    sudo mount -t msdos /dev/disk2 /Volumes/BOOTLOADER
fi
tinygo flash -target=max32620fthr .