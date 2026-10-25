//go:build tinygo && !qbs_noinput

package dev

import "time"

func waitGo(class string) {
	for {
		for Serial.Buffered() > 0 {
			b, _ := Serial.ReadByte()
			if b == 'G' {
				return
			}
		}
		println("@@READY qbs", class)
		time.Sleep(300 * time.Millisecond)
	}
}
