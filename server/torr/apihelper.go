package torr

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	sets "server/settings"
	"server/web/auth"
)

func getServer(user string) *BTServer {
	bt, err := ConnectServer(user)
	if err != nil {
		log.TLogln("error connect torrent server:", user, err)
		return nil
	}
	return bt
}

func LoadTorrent(user string, tor *Torrent) *Torrent {
	if tor.TorrentSpec == nil {
		return nil
	}
	bt := getServer(user)
	if bt == nil {
		return nil
	}
	tr, err := NewTorrent(tor.TorrentSpec, bt)
	if err != nil {
		return nil
	}
	if !tr.WaitInfo() {
		return nil
	}
	tr.Title = tor.Title
	tr.Poster = tor.Poster
	tr.Data = tor.Data
	return tr
}

func AddTorrent(user string, spec *torrent.TorrentSpec, title, poster string, data string, category string) (*Torrent, error) {
	bt := getServer(user)
	if bt == nil {
		return nil, errors.New("no torrent server")
	}

	torr, err := NewTorrent(spec, bt)
	if err != nil {
		log.TLogln("error add torrent:", user, err)
		return nil, err
	}

	torDB := GetTorrentDB(user, spec.InfoHash)

	if torr.Title == "" {
		torr.Title = title
		if title == "" && torDB != nil {
			torr.Title = torDB.Title
		}
		if torr.Title == "" && torr.Torrent != nil && torr.Torrent.Info() != nil {
			torr.Title = torr.Info().Name
		}
	}

	if torr.Category == "" {
		torr.Category = category
		if torr.Category == "" && torDB != nil {
			torr.Category = torDB.Category
		}
	}

	if torr.Poster == "" {
		torr.Poster = poster
		if torr.Poster == "" && torDB != nil {
			torr.Poster = torDB.Poster
		}
	}

	if torr.Data == "" {
		torr.Data = data
		if torr.Data == "" && torDB != nil {
			torr.Data = torDB.Data
		}
	}

	return torr, nil
}

func SaveTorrentToDB(user string, torr *Torrent) {
	log.TLogln("save to db:", user, torr.Hash())
	AddTorrentDB(user, torr)
}

func GetTorrent(user, hashHex string) *Torrent {
	if sets.HttpAuth {
		if user == "" {
			return nil
		}

		if !auth.UserExists(user) {
			return nil
		}
	}

	hash := metainfo.NewHashFromHex(hashHex)
	timeout := time.Second * time.Duration(sets.BTsets.TorrentDisconnectTimeout)
	if timeout > time.Minute {
		timeout = time.Minute
	}
	bt := getServer(user)
	if bt == nil {
		return nil
	}
	tor := bt.GetTorrent(hash)
	if tor != nil {
		tor.AddExpiredTime(timeout)
		return tor
	}

	tr := GetTorrentDB(user, hash)
	if tr != nil {
		tor = tr
		go func() {
			log.TLogln("New torrent", user, tor.Hash())
			tr, _ := NewTorrent(tor.TorrentSpec, bt)
			if tr != nil {
				tr.Title = tor.Title
				tr.Poster = tor.Poster
				tr.Data = tor.Data
				tr.Size = tor.Size
				tr.Timestamp = tor.Timestamp
				tr.Category = tor.Category
				tr.GotInfo()
			}
		}()
	}
	return tor
}

func SetTorrent(user, hashHex, title, poster, category string, data string) *Torrent {
	hash := metainfo.NewHashFromHex(hashHex)
	bt := getServer(user)
	if bt == nil {
		return nil
	}
	torr := bt.GetTorrent(hash)
	torrDb := GetTorrentDB(user, hash)

	if title == "" && torr == nil && torrDb != nil {
		torr = GetTorrent(user, hashHex)
		torr.GotInfo()
		if torr.Torrent != nil && torr.Torrent.Info() != nil {
			title = torr.Info().Name
		}
	}

	if torr != nil {
		if title == "" && torr.Torrent != nil && torr.Torrent.Info() != nil {
			title = torr.Info().Name
		}
		torr.Title = title
		torr.Poster = poster
		torr.Category = category
		if data != "" {
			torr.Data = data
		}
	}
	// update torrent data in DB
	if torrDb != nil {
		torrDb.Title = title
		torrDb.Poster = poster
		torrDb.Category = category
		if data != "" {
			torrDb.Data = data
		}
		AddTorrentDB(user, torrDb)
	}
	if torr != nil {
		return torr
	} else {
		return torrDb
	}
}

func RemTorrent(user, hashHex string) {
	if sets.ReadOnly {
		log.TLogln("API RemTorrent: Read-only DB mode!", user, hashHex)
		return
	}
	hash := metainfo.NewHashFromHex(hashHex)
	bt := getServer(user)
	if bt != nil && bt.RemoveTorrent(hash) {
		if sets.BTsets.UseDisk && hashHex != "" && hashHex != "/" {
			name := filepath.Join(sets.BTsets.TorrentsSavePath, hashHex)
			ff, _ := os.ReadDir(name)
			for _, f := range ff {
				os.Remove(filepath.Join(name, f.Name()))
			}
			err := os.Remove(name)
			if err != nil {
				log.TLogln("Error remove cache:", user, err)
			}
		}
	}
	RemTorrentDB(user, hash)
}

func ListTorrent(user string) []*Torrent {
	bt := getServer(user)
	if bt == nil {
		return nil
	}
	btlist := bt.ListTorrents()
	dblist := ListTorrentsDB(user)

	for hash, t := range dblist {
		if _, ok := btlist[hash]; !ok {
			btlist[hash] = t
		}
	}
	var ret []*Torrent

	for _, t := range btlist {
		ret = append(ret, t)
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Timestamp != ret[j].Timestamp {
			return ret[i].Timestamp > ret[j].Timestamp
		} else {
			return ret[i].Title > ret[j].Title
		}
	})

	return ret
}

func ListAllTorrents() []*Torrent {
	dblist := ListAllTorrentsDB()
	btlist := make(map[metainfo.Hash]*Torrent)

	ForEachServer(func(user string, bt *BTServer) {
		for hash, t := range bt.ListTorrents() {
			btlist[hash] = t
		}
		for hash, t := range ListTorrentsDB(user) {
			if _, ok := btlist[hash]; !ok {
				btlist[hash] = t
			}
		}
	})

	for hash, t := range dblist {
		if _, ok := btlist[hash]; !ok {
			btlist[hash] = t
		}
	}

	var ret []*Torrent
	for _, t := range btlist {
		ret = append(ret, t)
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Timestamp != ret[j].Timestamp {
			return ret[i].Timestamp > ret[j].Timestamp
		}
		return ret[i].Title > ret[j].Title
	})

	return ret
}

func DropTorrent(user, hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	bt := getServer(user)
	if bt == nil {
		return
	}
	bt.RemoveTorrent(hash)
}

func SetSettings(set *sets.BTSets) {
	if sets.ReadOnly {
		log.TLogln("API SetSettings: Read-only DB mode!")
		return
	}
	sets.SetBTSets(set)
	restartServers()
}

func SetDefSettings() {
	if sets.ReadOnly {
		log.TLogln("API SetDefSettings: Read-only DB mode!")
		return
	}
	sets.SetDefaultConfig()
	restartServers()
}

func dropAllTorrent(bt *BTServer) {
	for _, torr := range bt.torrents {
		torr.drop()
		<-torr.closed
	}
}

func restartServers() {
	log.TLogln("drop all torrents")
	ForEachServer(func(user string, bt *BTServer) {
		dropAllTorrent(bt)
	})
	time.Sleep(time.Second * 1)
	log.TLogln("disconect")
	ForEachServer(func(user string, bt *BTServer) {
		bt.Disconnect()
	})
	log.TLogln("connect")
	ForEachServer(func(user string, bt *BTServer) {
		if _, err := ConnectServer(user); err != nil {
			log.TLogln("error reconnect torrent server", user, err)
		}
	})
	time.Sleep(time.Second * 1)
	log.TLogln("end set settings")
}

func WriteStatus(user string, w io.Writer) {
	bt := getServer(user)
	if bt == nil || bt.client == nil {
		return
	}
	bt.client.WriteStatus(w)
}

func Preload(torr *Torrent, index int) {
	cache := float32(sets.BTsets.CacheSize)
	preload := float32(sets.BTsets.PreloadCache)
	size := int64((cache / 100.0) * preload)
	if size <= 0 {
		return
	}
	if size > sets.BTsets.CacheSize {
		size = sets.BTsets.CacheSize
	}
	torr.Preload(index, size)
}
