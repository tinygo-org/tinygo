package main

type sample struct {
	name    string
	value   int32
	flags   byte
	payload []byte
}

type decoder interface {
	decode([]byte) (sample, bool)
}

type temperatureDecoder struct{}
type counterDecoder struct{}
type textDecoder struct{}
type commandDecoder struct{}
type identityDecoder struct{}

func (temperatureDecoder) decode(packet []byte) (sample, bool) {
	if len(packet) < 4 {
		return sample{}, false
	}
	value := int32(packet[1])<<8 | int32(packet[2])
	if packet[1]&0x80 != 0 {
		value -= 1 << 16
	}
	return sample{
		name:    "temperature",
		value:   value,
		flags:   packet[3],
		payload: packet[4:],
	}, true
}

func (counterDecoder) decode(packet []byte) (sample, bool) {
	if len(packet) < 6 {
		return sample{}, false
	}
	value := int32(packet[1]) |
		int32(packet[2])<<8 |
		int32(packet[3])<<16 |
		int32(packet[4])<<24
	return sample{
		name:    "counter",
		value:   value,
		flags:   packet[5],
		payload: packet[6:],
	}, true
}

func (textDecoder) decode(packet []byte) (sample, bool) {
	if len(packet) < 3 {
		return sample{}, false
	}
	size := int(packet[1])
	if size > len(packet)-3 {
		return sample{}, false
	}
	payload := append([]byte(nil), packet[2:2+size]...)
	return sample{
		name:    string(payload),
		value:   int32(size),
		flags:   packet[2+size],
		payload: payload,
	}, true
}

func (commandDecoder) decode(packet []byte) (sample, bool) {
	if len(packet) < 6 || string(packet[1:5]) != "sync" {
		return sample{}, false
	}
	return sample{
		name:  "command-" + string(packet[5:6]),
		value: int32(packet[5]),
	}, true
}

type deviceID [8]byte

var trustedDevice = deviceID{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe}

func (identityDecoder) decode(packet []byte) (sample, bool) {
	if len(packet) < 10 {
		return sample{}, false
	}
	var id deviceID
	copy(id[:], packet[1:9])
	if id != trustedDevice {
		return sample{}, false
	}
	return sample{
		name:  "identity",
		value: int32(packet[9]),
	}, true
}

type transform interface {
	apply(sample) (sample, bool)
}

type clamp struct {
	min int32
	max int32
}

func (c clamp) apply(value sample) (sample, bool) {
	if value.value < c.min {
		value.value = c.min
		value.flags |= 1
	} else if value.value > c.max {
		value.value = c.max
		value.flags |= 2
	}
	return value, true
}

type scale struct {
	numerator   int32
	denominator int32
}

func (s scale) apply(value sample) (sample, bool) {
	if s.denominator == 0 {
		return sample{}, false
	}
	value.value = value.value * s.numerator / s.denominator
	return value, true
}

type deduplicate struct {
	previous map[string]int32
}

func (d *deduplicate) apply(value sample) (sample, bool) {
	previous, ok := d.previous[value.name]
	d.previous[value.name] = value.value
	return value, !ok || previous != value.value
}

type pipeline struct {
	transforms []transform
}

func (p pipeline) process(value sample) (sample, bool) {
	for _, transform := range p.transforms {
		var ok bool
		value, ok = transform.apply(value)
		if !ok {
			return sample{}, false
		}
	}
	return value, true
}

type queue[T any] struct {
	values []T
	head   int
}

func (q *queue[T]) push(value T) {
	q.values = append(q.values, value)
}

func (q *queue[T]) pop() (T, bool) {
	if q.head == len(q.values) {
		var zero T
		q.values = q.values[:0]
		q.head = 0
		return zero, false
	}
	value := q.values[q.head]
	q.head++
	return value, true
}

type statistic struct {
	count int32
	sum   int64
	min   int32
	max   int32
}

func (s *statistic) add(value int32) {
	if s.count == 0 || value < s.min {
		s.min = value
	}
	if s.count == 0 || value > s.max {
		s.max = value
	}
	s.count++
	s.sum += int64(value)
}

func (s statistic) average() int32 {
	if s.count == 0 {
		return 0
	}
	return int32(s.sum / int64(s.count))
}

type collector struct {
	statistics map[string]*statistic
	output     []byte
}

func (c *collector) consume(value sample) {
	stats := c.statistics[value.name]
	if stats == nil {
		stats = &statistic{}
		c.statistics[value.name] = stats
	}
	stats.add(value.value)
	c.output = appendRecord(c.output, value, *stats)
}

func appendRecord(dst []byte, value sample, stats statistic) []byte {
	dst = append(dst, '{')
	dst = appendString(dst, "name")
	dst = append(dst, ':')
	dst = appendString(dst, value.name)
	dst = append(dst, ',')
	dst = appendString(dst, "value")
	dst = append(dst, ':')
	dst = appendInt(dst, value.value)
	dst = append(dst, ',')
	dst = appendString(dst, "average")
	dst = append(dst, ':')
	dst = appendInt(dst, stats.average())
	dst = append(dst, ',')
	dst = appendString(dst, "range")
	dst = append(dst, ':', '[')
	dst = appendInt(dst, stats.min)
	dst = append(dst, ',')
	dst = appendInt(dst, stats.max)
	dst = append(dst, ']', ',')
	dst = appendString(dst, "flags")
	dst = append(dst, ':')
	dst = appendUint(dst, uint32(value.flags))
	dst = append(dst, '}', '\n')
	return dst
}

func appendString(dst []byte, value string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '"', '\\':
			dst = append(dst, '\\', value[i])
		case '\n':
			dst = append(dst, '\\', 'n')
		default:
			dst = append(dst, value[i])
		}
	}
	return append(dst, '"')
}

func appendInt(dst []byte, value int32) []byte {
	if value < 0 {
		dst = append(dst, '-')
		return appendUint(dst, uint32(-value))
	}
	return appendUint(dst, uint32(value))
}

func appendUint(dst []byte, value uint32) []byte {
	if value >= 10 {
		dst = appendUint(dst, value/10)
	}
	return append(dst, byte(value%10)+'0')
}

type packetStream struct {
	data   []byte
	offset int
}

func (s *packetStream) next() ([]byte, bool) {
	if s.offset >= len(s.data) {
		return nil, false
	}
	size := int(s.data[s.offset])
	s.offset++
	if size > len(s.data)-s.offset {
		s.offset = len(s.data)
		return nil, false
	}
	packet := s.data[s.offset : s.offset+size]
	s.offset += size
	return packet, true
}

type handler func(sample) bool

func dispatch(value sample, handlers map[byte]handler) bool {
	if handler := handlers[value.flags&3]; handler != nil {
		return handler(value)
	}
	return false
}

func relay(input <-chan sample, output chan<- sample) {
	for value := range input {
		output <- value
	}
	close(output)
}

var input = []byte{
	6, 1, 0, 25, 0, 1, 2,
	8, 2, 42, 0, 0, 0, 2, 3, 4,
	8, 3, 4, 'f', 'a', 'n', '1', 3, 9,
	6, 4, 's', 'y', 'n', 'c', 7,
	10, 5, 0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe, 9,
	6, 1, 0, 26, 0, 1, 2,
	8, 2, 43, 0, 0, 0, 2, 3, 4,
}

var result []byte

func main() {
	decoders := map[byte]decoder{
		1: temperatureDecoder{},
		2: counterDecoder{},
		3: textDecoder{},
		4: commandDecoder{},
		5: identityDecoder{},
	}
	processor := pipeline{transforms: []transform{
		clamp{min: -4000, max: 12000},
		scale{numerator: 10, denominator: 1},
		&deduplicate{previous: make(map[string]int32)},
	}}
	var pending queue[sample]
	stream := packetStream{data: input}
	for {
		packet, ok := stream.next()
		if !ok {
			break
		}
		if len(packet) == 0 {
			continue
		}
		decoder := decoders[packet[0]]
		if decoder == nil {
			continue
		}
		value, ok := decoder.decode(packet)
		if !ok {
			continue
		}
		value, ok = processor.process(value)
		if ok {
			pending.push(value)
		}
	}

	sink := collector{statistics: make(map[string]*statistic)}
	handlers := map[byte]handler{
		0: func(value sample) bool {
			sink.consume(value)
			return true
		},
		1: func(value sample) bool {
			value.name += "-alert"
			sink.consume(value)
			return true
		},
		2: func(value sample) bool {
			value.value = -value.value
			sink.consume(value)
			return true
		},
	}
	input := make(chan sample, 8)
	output := make(chan sample, 8)
	go relay(input, output)
	for {
		value, ok := pending.pop()
		if !ok {
			break
		}
		input <- value
	}
	close(input)
	for value := range output {
		dispatch(value, handlers)
	}
	result = sink.output
	println(len(result))
}
