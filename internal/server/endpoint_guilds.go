package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
)

// (GET /v1/guilds/{guild_ident}/users)
func (e *Endpoints) GetGuildUsers(c *gin.Context, guildIdent api.GuildIdent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	guildID := guildIdent

	var users []api.User
	err := orm.DB().NewSelect().
		Model(&users).
		Join("INNER JOIN accounts as account ON \"user\".id = account.user_id").
		Relation("Bans").
		Relation("Accounts").
		Relation("PlatformLinks").
		Where("CAST(\"account\".\"guilds\" AS text) LIKE ?", "%"+guildID+"%").
		Scan(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &users)
}
