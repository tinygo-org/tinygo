package cases

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
	"strings"
	"unicode/utf8"
)

func init() {
	register("crc32", 0, func() string {
		data := randBytes(1500, 10)
		castagnoli := crc32.MakeTable(crc32.Castagnoli)
		sum := crc32.Update(0, castagnoli, data[:700])
		sum = crc32.Update(sum, castagnoli, data[700:])
		return hex64(uint64(crc32.ChecksumIEEE(data))) + " " + hex64(uint64(sum))
	})
	// Two slicing tables of 8KB each.
	needsHeap(20 << 10)

	register("adler32-fnv", 0, func() string {
		data := randBytes(1000, 11)
		h := fnv.New64a()
		h.Write(data)
		h32 := fnv.New32()
		h32.Write(data)
		return hex64(uint64(adler32.Checksum(data))) + " " + hex64(h.Sum64()) + " " + hex64(uint64(h32.Sum32()))
	})

	register("sha256", 0, func() string {
		sum := sha256.Sum256(randBytes(1000, 12))
		return hex.EncodeToString(sum[:])
	})
	needsHeap(3 << 10)

	register("sha1-md5", 0, func() string {
		data := randBytes(777, 13)
		s1 := sha1.Sum(data)
		m5 := md5.Sum(data)
		return hex.EncodeToString(s1[:]) + " " + hex.EncodeToString(m5[:])
	})

	register("encoding", 0, func() string {
		data := randBytes(300, 14)
		b64 := base64.StdEncoding.EncodeToString(data)
		back, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || string(back) != string(data) {
			return "base64 round trip failed"
		}
		raw := base64.RawURLEncoding.EncodeToString(data[:101])
		buf := make([]byte, 16)
		binary.LittleEndian.PutUint64(buf, binary.BigEndian.Uint64(data))
		binary.BigEndian.PutUint32(buf[8:], binary.LittleEndian.Uint32(data[8:]))
		return hex64(fnv64(b64)) + " " + hex64(fnv64(raw)) + " " + hex.EncodeToString(buf)
	})
	needsHeap(4 << 10)

	register("strings", 0, func() string {
		var sb strings.Builder
		r := rng(15)
		for i := 0; i < 40; i++ {
			sb.WriteRune(rune(0x20 + r.next()%0x2000))
			sb.WriteString(" ab,cd;")
		}
		s := sb.String()
		parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' })
		joined := strings.Join(parts, "|")
		up := strings.ToUpper(strings.Repeat(joined[:20], 3))
		rep := strings.ReplaceAll(s, "ab", "xyz")
		return hex64(fnv64(joined)) + " " + hex64(fnv64(up)) + " " + hex64(fnv64(rep)) +
			" " + hex64(uint64(utf8.RuneCountInString(s))) + " " + hex64(uint64(strings.Index(s, "cd;")))
	})
}
