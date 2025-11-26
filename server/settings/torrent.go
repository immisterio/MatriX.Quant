package settings

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentDB struct {
	*torrent.TorrentSpec

	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Poster   string `json:"poster,omitempty"`
	Data     string `json:"data,omitempty"`

	Timestamp int64 `json:"timestamp,omitempty"`
	Size      int64 `json:"size,omitempty"`
}

type File struct {
	Name string `json:"name,omitempty"`
	Id   int    `json:"id,omitempty"`
	Size int64  `json:"size,omitempty"`
}

var mu sync.Mutex

func AddTorrent(user string, torr *TorrentDB) {
	mu.Lock()
	xpath := joinUserXPath("Torrents", user)
	list := listTorrentLocked(xpath)
	if len(list) == 0 && normalizeUserID(user) == "base" {
		list = listTorrentLocked("Torrents")
	}
	find := -1
	for i, db := range list {
		if db.InfoHash.HexString() == torr.InfoHash.HexString() {
			find = i
			break
		}
	}
	if find != -1 {
		list[find] = torr
	} else {
		list = append(list, torr)
	}
	for _, db := range list {
		buf, err := json.Marshal(db)
		if err == nil {
			tdb.Set(xpath, db.InfoHash.HexString(), buf)
		}
	}
	mu.Unlock()
}

func ListTorrent(user string) []*TorrentDB {
	mu.Lock()
	defer mu.Unlock()

	list := listTorrentLocked(joinUserXPath("Torrents", user))
	if len(list) == 0 && normalizeUserID(user) == "base" {
		return listTorrentLocked("Torrents")
	}
	return list
}

func RemTorrent(user string, hash metainfo.Hash) {
	mu.Lock()
	tdb.Rem(joinUserXPath("Torrents", user), hash.HexString())
	mu.Unlock()
}

func listTorrentLocked(xpath string) []*TorrentDB {
	var list []*TorrentDB
	keys := tdb.List(xpath)
	for _, key := range keys {
		buf := tdb.Get(xpath, key)
		if len(buf) > 0 {
			var torr *TorrentDB
			err := json.Unmarshal(buf, &torr)
			if err == nil {
				list = append(list, torr)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp > list[j].Timestamp
	})
	return list
}

func ListAllTorrents() []*TorrentDB {
	mu.Lock()
	defer mu.Unlock()

	var list []*TorrentDB
	seen := make(map[string]struct{})
	for _, tor := range listTorrentLocked("Torrents") {
		key := tor.InfoHash.HexString()
		seen[key] = struct{}{}
		list = append(list, tor)
	}
	users := tdb.List("Torrents")
	for _, user := range users {
		for _, tor := range listTorrentLocked(joinUserXPath("Torrents", user)) {
			key := tor.InfoHash.HexString()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			list = append(list, tor)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp > list[j].Timestamp
	})

	return list
}
