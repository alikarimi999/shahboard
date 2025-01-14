package http

import (
	"net/http"

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
	pgn, err := r.s.GetGamePGN(ctx, oid)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, getGamePGNResponse{ID: oid, PGN: pgn})
}

func (r *Router) getLiveGames(ctx *gin.Context) {
	games, err := r.s.GetLiveGames(ctx)
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

func (r *Router) getGamesFen(ctx *gin.Context) {
	req := &getGamesFENRequest{}
	if err := ctx.BindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	res, err := r.s.GetGamesFEN(ctx, req.Games)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	fens := make([]fen, 0, len(res))
	for id, f := range res {
		fens = append(fens, fen{ID: id, FEN: f})
	}

	ctx.JSON(http.StatusOK, list{
		List: []interface{}{
			fens,
		},
	})

	ctx.JSON(http.StatusOK, fens)
}
