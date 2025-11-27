package api

import (
	"net/http"
	"strings"

	"server/log"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Action: add, get, set, rem, list, drop
type torrReqJS struct {
	requestI
	Link     string `json:"link,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Poster   string `json:"poster,omitempty"`
	Data     string `json:"data,omitempty"`
	SaveToDB bool   `json:"save_to_db,omitempty"`
}

// torrents godoc
//
//	@Summary		Handle torrents informations
//	@Description	Allow to list, add, remove, get, set, drop, wipe torrents on server. The action depends of what has been asked.
//
//	@Tags			API
//
//	@Param			request	body	torrReqJS	true	"Torrent request. Available params for action: add, get, set, rem, list, drop, wipe. link required for add, hash required for get, set, rem, drop."
//
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/torrents [post]
func torrents(c *gin.Context) {
	var req torrReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	user := utils.UserID(c)
	c.Status(http.StatusBadRequest)
	switch req.Action {
	case "add":
		{
			addTorrent(user, req, c)
		}
	case "get":
		{
			getTorrent(user, req, c)
		}
	case "set":
		{
			setTorrent(user, req, c)
		}
	case "rem":
		{
			remTorrent(user, req, c)
		}
	case "list":
		{
			listTorrents(user, c)
		}
	case "drop":
		{
			dropTorrent(user, req, c)
		}
	case "wipe":
		{
			wipeTorrents(user, c)
		}
	}
}

func addTorrent(user string, req torrReqJS, c *gin.Context) {
	if req.Link == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("link is empty"))
		return
	}

	log.TLogln("add torrent", user, req.Link)
	req.Link = strings.ReplaceAll(req.Link, "&amp;", "&")
	torrSpec, err := utils.ParseLink(req.Link)
	if err != nil {
		log.TLogln("error parse link:", user, err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tor, err := torr.AddTorrent(user, torrSpec, req.Title, req.Poster, req.Data, req.Category)
	// if tor.Data != "" && set.BTsets.EnableDebug {
	// 	log.TLogln("torrent data:", tor.Data)
	// }
	// if tor.Category != "" && set.BTsets.EnableDebug {
	// 	log.TLogln("torrent category:", tor.Category)
	// }
	if err != nil {
		log.TLogln("error add torrent:", user, err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	go func() {
		if !tor.GotInfo() {
			log.TLogln("error add torrent:", user, "timeout connection get torrent info")
			return
		}

		if tor.Title == "" {
			tor.Title = torrSpec.DisplayName // prefer dn over name
			tor.Title = strings.ReplaceAll(tor.Title, "rutor.info", "")
			tor.Title = strings.ReplaceAll(tor.Title, "_", " ")
			tor.Title = strings.Trim(tor.Title, " ")
			if tor.Title == "" {
				tor.Title = tor.Name()
			}
		}

		if req.SaveToDB {
			torr.SaveTorrentToDB(user, tor)
		}
	}()

	st := tor.Status()
	st.Hash = utils.JoinHashUser(st.Hash, user)
	c.JSON(200, st)
}

func getTorrent(user string, req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	tor := torr.GetTorrent(reqUser, hash)

	if tor != nil {
		st := tor.Status()
		st.Hash = utils.JoinHashUser(st.Hash, reqUser)
		c.JSON(200, st)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func setTorrent(user string, req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	torr.SetTorrent(reqUser, hash, req.Title, req.Poster, req.Category, req.Data)
	c.Status(200)
}

func remTorrent(user string, req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	torr.RemTorrent(reqUser, hash)
	c.Status(200)
}

func listTorrents(user string, c *gin.Context) {
	list := torr.ListTorrent(user)
	if len(list) == 0 {
		c.JSON(200, []*state.TorrentStatus{})
		return
	}
	var stats []*state.TorrentStatus
	for _, tr := range list {
		st := tr.Status()
		st.Hash = utils.JoinHashUser(st.Hash, user)
		stats = append(stats, st)
	}
	c.JSON(200, stats)
}

func dropTorrent(user string, req torrReqJS, c *gin.Context) {
	if req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	torr.DropTorrent(reqUser, hash)
	c.Status(200)
}

func wipeTorrents(user string, c *gin.Context) {
	torrents := torr.ListTorrent(user)
	for _, t := range torrents {
		torr.RemTorrent(user, t.TorrentSpec.InfoHash.HexString())
	}
	c.Status(200)
}
