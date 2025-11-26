package settings

import (
	"encoding/json"

	"server/log"
)

type Viewed struct {
	Hash      string `json:"hash"`
	FileIndex int    `json:"file_index"`
}

func SetViewed(user string, vv *Viewed) {
	var indexes map[string]map[int]struct{}
	var err error

	path := "Viewed"
	buf := tdb.Get(path, vv.Hash)
	if len(buf) == 0 {
		indexes = make(map[string]map[int]struct{})
	} else {
		err = json.Unmarshal(buf, &indexes)
		if err != nil {
			log.TLogln("Error decode viewed:", user, err)
			return
		}
	}

	userKey := normalizeUserID(user)
	if _, ok := indexes[userKey]; !ok {
		indexes[userKey] = make(map[int]struct{})
	}
	indexes[userKey][vv.FileIndex] = struct{}{}

	buf, err = json.Marshal(indexes)
	if err == nil {
		tdb.Set(path, vv.Hash, buf)
	} else {
		log.TLogln("Error set viewed:", user, err)
	}
}

func RemViewed(user string, vv *Viewed) {
	path := "Viewed"
	buf := tdb.Get(path, vv.Hash)
	var indeces map[string]map[int]struct{}
	err := json.Unmarshal(buf, &indeces)
	if err == nil {
		userKey := normalizeUserID(user)
		if vv.FileIndex != -1 {
			if entries, ok := indeces[userKey]; ok {
				delete(entries, vv.FileIndex)
				if len(entries) == 0 {
					delete(indeces, userKey)
				}
				if len(indeces) == 0 {
					tdb.Rem(path, vv.Hash)
					return
				}
			}

			buf, err = json.Marshal(indeces)
			if err == nil {
				tdb.Set(path, vv.Hash, buf)
			}
		} else {
			delete(indeces, userKey)
			if len(indeces) == 0 {
				tdb.Rem(path, vv.Hash)
				return
			}
			buf, err = json.Marshal(indeces)
			if err == nil {
				tdb.Set(path, vv.Hash, buf)
			}
		}
	}
	if err != nil {
		log.TLogln("Error rem viewed:", user, err)
	}
}

func ListViewed(hash string, user string) []*Viewed {
	var err error
	path := "Viewed"
	userKey := normalizeUserID(user)
	if hash != "" {
		buf := tdb.Get(path, hash)
		if len(buf) == 0 {
			return []*Viewed{}
		}
		var indeces map[string]map[int]struct{}
		err = json.Unmarshal(buf, &indeces)
		if err == nil {
			var ret []*Viewed
			for i := range indeces[userKey] {
				ret = append(ret, &Viewed{hash, i})
			}
			return ret
		}
	} else {
		var ret []*Viewed
		keys := tdb.List(path)
		for _, key := range keys {
			buf := tdb.Get(path, key)
			if len(buf) == 0 {
				return []*Viewed{}
			}
			var indeces map[string]map[int]struct{}
			err = json.Unmarshal(buf, &indeces)
			if err == nil {
				for i := range indeces[userKey] {
					ret = append(ret, &Viewed{key, i})
				}
			}
		}
		return ret
	}

	log.TLogln("Error list viewed:", user, err)
	return []*Viewed{}
}
