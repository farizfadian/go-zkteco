package protocol

import (
	"encoding/binary"
	"time"
)

// ZKTeco uses a custom epoch: 2000-01-01 00:00:00 local time
var zkEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)

// EncodeTime encodes a time.Time to ZKTeco format (seconds since 2000-01-01).
func EncodeTime(t time.Time) uint32 {
	return uint32(t.Sub(zkEpoch).Seconds())
}

// DecodeTime decodes a ZKTeco timestamp to time.Time.
func DecodeTime(seconds uint32) time.Time {
	return zkEpoch.Add(time.Duration(seconds) * time.Second)
}

// EncodeTimeBytes encodes a time.Time to 4 bytes (little-endian).
func EncodeTimeBytes(t time.Time) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, EncodeTime(t))
	return buf
}

// DecodeTimeBytes decodes 4 bytes to time.Time.
func DecodeTimeBytes(data []byte) time.Time {
	if len(data) < 4 {
		return time.Time{}
	}
	seconds := binary.LittleEndian.Uint32(data)
	return DecodeTime(seconds)
}

// DecodeDateTimeBytes decodes the alternative datetime format used in some responses.
// Format: year(1) + month(1) + day(1) + hour(1) + minute(1) + second(1)
// Year is offset from 2000
func DecodeDateTimeBytes(data []byte) time.Time {
	if len(data) < 6 {
		return time.Time{}
	}

	year := 2000 + int(data[0])
	month := time.Month(data[1])
	day := int(data[2])
	hour := int(data[3])
	minute := int(data[4])
	second := int(data[5])

	return time.Date(year, month, day, hour, minute, second, 0, time.Local)
}

// EncodeDateTimeBytes encodes a time.Time to 6 bytes.
func EncodeDateTimeBytes(t time.Time) []byte {
	return []byte{
		byte(t.Year() - 2000),
		byte(t.Month()),
		byte(t.Day()),
		byte(t.Hour()),
		byte(t.Minute()),
		byte(t.Second()),
	}
}

// ParseAttendanceTime parses the packed time format used in attendance records.
// The format packs datetime into a single uint32:
// bits 0-5: second/2
// bits 6-11: minute
// bits 12-16: hour
// bits 17-21: day
// bits 22-25: month
// bits 26-31: year (offset from 2000)
func ParseAttendanceTime(packed uint32) time.Time {
	second := int((packed & 0x3F) * 2)
	minute := int((packed >> 6) & 0x3F)
	hour := int((packed >> 12) & 0x1F)
	day := int((packed >> 17) & 0x1F)
	month := time.Month((packed >> 22) & 0x0F)
	year := 2000 + int((packed>>26)&0x3F)

	return time.Date(year, month, day, hour, minute, second, 0, time.Local)
}

// PackAttendanceTime packs a time.Time into the attendance record format.
func PackAttendanceTime(t time.Time) uint32 {
	var packed uint32
	packed |= uint32(t.Second() / 2)
	packed |= uint32(t.Minute()) << 6
	packed |= uint32(t.Hour()) << 12
	packed |= uint32(t.Day()) << 17
	packed |= uint32(t.Month()) << 22
	packed |= uint32(t.Year()-2000) << 26
	return packed
}
