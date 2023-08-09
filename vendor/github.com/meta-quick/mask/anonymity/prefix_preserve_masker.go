package anonymity

import (
	"fmt"
	"github.com/meta-quick/mask/utils"
	"strconv"
)

type PrefixPreserveMasker struct {
}

func NewPrefixPreserveMasker() *PrefixPreserveMasker {
	return &PrefixPreserveMasker{}
}

func (m *PrefixPreserveMasker) mapping(c rune,fixed bool) rune {
	return utils.CharMapping(c,'*',fixed)
}

func (p PrefixPreserveMasker) MaskInteger(i int64,preserved int) (int64,error) {
	orgin := fmt.Sprintf("%d",i)
	var r = []rune(orgin)

	for i := preserved; i < len(r); i++ {
		r[i] = p.mapping(r[i],false)
	}
	return strconv.ParseInt(string(r),10,64)
}

func (p PrefixPreserveMasker) MaskString(in string,preserved int) (string,error) {
	var r = []rune(in)

	for i := preserved; i < len(r); i++ {
		r[i] = p.mapping(r[i],true)
	}

	return string(r),nil
}