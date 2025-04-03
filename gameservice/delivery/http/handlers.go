package http

import (
	"net/http"

	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (r *Router) getGamePGN(ctx *gin.Context) {
	sid := ctx.Param("id")

	oid, err := types.ParseObjectId(sid)
	if err != nil {
		ctx.JSON(400, err)
		return
	}
	res, err := r.s.GetGamePGN(ctx, oid)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, res)
}

func (r *Router) getLiveGames(ctx *gin.Context) {
	// check userId query
	userId := ctx.Query("user_id")
	if userId != "" {
		uid, err := types.ParseObjectId(userId)
		if err != nil {
			ctx.JSON(400, err)
			return
		}

		res, err := r.s.GetLiveGamePgnByUserID(ctx, uid)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(200, res)
		return
	}

	gameId := ctx.Query("game_id")
	if gameId != "" {
		gid, err := types.ParseObjectId(gameId)
		if err != nil {
			ctx.JSON(400, err)
			return
		}
		res, err := r.s.GetLiveGameByID(ctx, gid)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(200, res)
		return
	}

	games, err := r.s.GetLiveGamesIDs(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, list{
		List: []interface{}{
			games,
		},
	})
}

func (r *Router) getLiveGamesData(ctx *gin.Context) {
	games, err := r.s.GetLiveGamesData(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, list{
		List: []interface{}{
			games,
		},
	})
}

func (r *Router) getLiveGameByUserId(ctx *gin.Context) {
	userID := ctx.Param("id")
	uid, err := types.ParseObjectId(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	res, err := r.s.GetLiveGameIdByUserId(ctx, uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, GetLiveGameIdByUserIdRequest{GameId: res})
}

func (r *Router) resignByPlayer(ctx *gin.Context) {
	gameId := ctx.Param("gameId")
	gid, err := types.ParseObjectId(gameId)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	userID, ok := middleware.ExtractUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = r.s.ResingByPlayer(ctx, gid, userID.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, "ok")
}

// func (r *Router) getGamesFen(ctx *gin.Context) {
// 	req := &getGamesFENRequest{}
// 	if err := ctx.BindJSON(req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, err)
// 		return
// 	}

// 	res, err := r.s.GetGamesFEN(ctx, req.Games)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, err)
// 		return
// 	}

// 	fens := make([]fen, 0, len(res))
// 	for id, f := range res {
// 		fens = append(fens, fen{ID: id, FEN: f})
// 	}

// 	ctx.JSON(http.StatusOK, list{
// 		List: []interface{}{
// 			fens,
// 		},
// 	})

// 	ctx.JSON(http.StatusOK, fens)
// }
