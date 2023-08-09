package anonymity

import (
	"strings"
	"time"
)

type FloorMasker struct {
}

func NewFloorMasker() *FloorMasker {
	return &FloorMasker{}
}

func (m *FloorMasker) MaskFloat64(in float64) (int64, error) {
	return int64(in), nil
}

func (h *FloorMasker) MaskFloat32(in float32) (int32, error) {
	return int32(in), nil
}

func (h *FloorMasker) MaskTime(in time.Time,level string) (time.Time, error) {
	var year = in.Year()
	var month = in.Month()
	var day = in.Day()
	var hour = in.Hour()
	var minute = in.Minute()
	var second = in.Second()

	if strings.Contains(level,"Y"){
		year = year / 10 * 10
	}
	if strings.Contains(level,"M"){
		month = month / 6 * 6 + 1
	}
	if strings.Contains(level,"D"){
		day = day/10 * 10 + 1
	}
	if strings.Contains(level,"H"){
		hour = hour/6 *6 + 1
	}
	if strings.Contains(level,"m"){
		minute = 0
	}
	if strings.Contains(level,"s"){
		second = 0
	}

	return time.Date(year,month,day,hour,minute,second,0,time.Local),nil
}