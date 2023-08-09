package jsonmask

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/mask/types"
	"strconv"
	"strings"
	"sync"
)


type ProcessHandle struct {
	Fn string
	Args []string
}

var defaultSensitiveData = map[string]ProcessHandle{
	"name": ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"surName" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"firstName" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"lastName" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"identification" :ProcessHandle{
		Fn: types.HIDE_MASK_STRX.Name,
		Args: []string{"10","16"},
	},
	"national" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"card" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"phone" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"phoneNo" :ProcessHandle{
		Fn: types.PFE_MASK_NUM.Name,
		Args: []string{"6"},
	},
	"number" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"fnumber" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"username" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"password" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"email" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
	"address" :ProcessHandle{
		Fn: types.PFE_MASK_STR.Name,
		Args: []string{"2"},
	},
}

type Handle interface {
	Process(b []byte) (*string, error)
}

type Masker struct {
	sensitiveField map[string]ProcessHandle
}

func NewMasker() Masker {
	return Masker{
	}
}

func (m *Masker) Init(handles ...map[string]ProcessHandle) {
	var f = defaultSensitiveData
	for _, h := range handles {
		for k, v := range h {
			f[k] = v
		}
	}

	m.sensitiveField = f
}

func (m *Masker) walkThrough(b []byte, storage *[]types.BuiltinContext, p chan bool, wg *sync.WaitGroup) error {
	defer func() {
		<-p
		wg.Done()
	}()
	var gson interface{}
	err := sonic.Unmarshal(b, &gson)
	if err != nil {
		return err
	}
	switch t := gson.(type) {
	case map[string]interface{}:
		for k, v := range t {
			switch v := v.(type) {
			case string:
				m.sensitive(k, v, storage)
			case float64:
				m.sensitive(k, v, storage)
			case int32:
				m.sensitive(k, v, storage)
			case bool:
				m.sensitive(k, v, storage)
			case []interface{}:
				for _, val := range v {
					err = m.next(val, storage, p, wg)
					if err != nil {
						return err
					}
				}
			case map[string]interface{}:
				err = m.next(v, storage, p, wg)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *Masker) next(v interface{}, storage *[]types.BuiltinContext, p chan bool, wg *sync.WaitGroup) error {
	b, err := sonic.Marshal(v)
	if err != nil {
		return err
	}
	wg.Add(1)
	p <- true
	go m.walkThrough(b, storage, p, wg)
	return nil
}

func (m *Masker) sensitive(k string, v interface{}, storage *[]types.BuiltinContext) {
	for keyword := range m.sensitiveField {
		if strings.EqualFold(keyword,k) {
			ctx := types.BuiltinContext{
				Current: typeCasting(v),
				Fn: m.sensitiveField[keyword].Fn,
				Args: m.sensitiveField[keyword].Args,
			}
			*storage = append(*storage, ctx)
		}
	}
}

func (m *Masker) Process(b []byte) (*string, error) {
	var storage []types.BuiltinContext
	p := make(chan bool, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	p <- true
	err := m.walkThrough(b, &storage, p, &wg)
	if err != nil {
		return nil, err
	}
	wg.Wait()
	return masking(b, storage)
}

func masking(j []byte, d []types.BuiltinContext) (*string, error) {
	body := string(j)
	if len(d) == 0 {
		return &body, nil
	}

	for _, ctx := range d {
		types.Eval(&ctx)
		body = strings.ReplaceAll(body, ctx.Current,typeCasting(ctx.Result) )
	}
	return &body, nil
}

func typeCasting(d interface{}) string {
	switch c := d.(type) {
	case string:
		return c
	case int64:
		return fmt.Sprint(c)
	case float64:
		return strconv.FormatFloat(c, 'f', -1, 64)
	default:
		return fmt.Sprint(d)
	}
}