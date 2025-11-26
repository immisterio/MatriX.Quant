package api

import (
	"net/http"

	sets "server/settings"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

/*
file index starts from 1
*/

// Action: set, rem, list
type viewedReqJS struct {
	requestI
	*sets.Viewed
}

// viewed godoc
//
//	@Summary		Set / List / Remove viewed torrents
//	@Description	Allow to set, list or remove viewed torrents from server.
//
//	@Tags			API
//
//	@Param			request	body	viewedReqJS	true	"Viewed torrent request. Available params for action: set, rem, list"
//
//	@Accept			json
//	@Produce		json
//	@Success		200 {array} sets.Viewed
//	@Router			/viewed [post]
func viewed(c *gin.Context) {
	var req viewedReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch req.Action {
	case "set":
		{
			setViewed(req, c)
		}
	case "rem":
		{
			remViewed(req, c)
		}
	case "list":
		{
			listViewed(req, c)
		}
	}
}

func setViewed(req viewedReqJS, c *gin.Context) {
	if req.Viewed == nil || req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	user := utils.UserID(c)
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	req.Viewed.Hash = hash
	sets.SetViewed(reqUser, req.Viewed)
	c.Status(200)
}

func remViewed(req viewedReqJS, c *gin.Context) {
	if req.Viewed == nil || req.Hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}
	user := utils.UserID(c)
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	req.Viewed.Hash = hash
	sets.RemViewed(reqUser, req.Viewed)
	c.Status(200)
}

func listViewed(req viewedReqJS, c *gin.Context) {
	user := utils.UserID(c)
	hash, reqUser, ok := utils.ResolveHashUser(c, req.Hash, user)
	if !ok {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	list := sets.ListViewed(hash, reqUser)
	for _, v := range list {
		v.Hash = utils.JoinHashUser(v.Hash, reqUser)
	}
	c.JSON(200, list)
}
