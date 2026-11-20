//go:build boardpackage

package machine

func boardUSBIdentity() (uint16, uint16, string, string)

func defaultUSBIdentity() (vid, pid uint16, manufacturer, product string) {
	return boardUSBIdentity()
}
