package anonymity

import (
	"github.com/meta-quick/mask/utils"
	"time"
)

type HidingMasker struct {
}

func NewHidingMasker() *HidingMasker {
	return &HidingMasker{}
}

func (h *HidingMasker) MaskBool(in bool) (bool, error) {
	return true, nil
}

func (h *HidingMasker) MaskInt(in int) (int, error) {
	return 0, nil
}

func (h *HidingMasker) MaskInt64(in int64) (int64, error) {
	return 0, nil
}

func (h *HidingMasker) MaskUint(in uint) (uint, error) {
	return 0, nil
}

func (h *HidingMasker) MaskUint64(in uint64) (uint64, error) {
	return 0, nil
}

func (h *HidingMasker) MaskFloat32(in float32) (float32, error) {
	return 0, nil
}

func (h *HidingMasker) MaskFloat64(in float64) (float64, error) {
	return 0, nil
}

func (h *HidingMasker) MaskString(in string,replace string) (string, error) {
	return replace, nil
}

func (h *HidingMasker) MaskString0(in string,overlay string,start,end int) (string, error) {
	return utils.OverlayString(in,overlay,start,end),nil
}

func (h *HidingMasker) MaskTime(in *time.Time) (time.Time, error) {
	return time.Now(), nil
}

