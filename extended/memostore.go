package extended

import (
	"sync"
)

var (
	smKeyMap = sync.Map{}
)

func StoreAdd(ns string, key string, value string) {
	key1 := ns + "/" + key
	smKeyMap.Store(key1, value)
}

func StoreGet(ns string, key string) string {
	key1 := ns + "/" + key

	v, ok := smKeyMap.Load(key1)
	if ok {
		return v.(string)
	}
	return ""
}

func StoreDelete(ns string, key string) {
	key1 := ns + "/" + key
	smKeyMap.Delete(key1)
}

func SmKeyAdd(key string, value string) {
	StoreAdd("sm", key, value)
}

func SmKeyGet(key string) string {
	return StoreGet("ns", key)
}

func SmKeyDelete(key string) {
	StoreDelete("ns", key)
}
