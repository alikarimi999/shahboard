package http

import (
	"strconv"

	"github.com/alikarimi999/shahboard/pkg/paginate"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) setupRatingRoutes() {
	r := h.Group("/rating")
	r.GET("/:userId", h.getUserRating)
	r.GET("/history/:userId", h.getUserRatingHistory)
}
func (h *Handler) getUserRating(c *gin.Context) {
	sid := c.Param("userId")

	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	rating, err := h.rating.GetUserRating(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, UserRatingResponse{
		CurrentScore: rating.CurrentScore,
		BestScore:    rating.BestScore,
		GamesPlayed:  rating.GamesPlayed,
		GamesWon:     rating.GamesWon,
		GamesLost:    rating.GamesLost,
		GamesDraw:    rating.GamesDraw,
		LastUpdated:  rating.LastUpdated.Unix(),
	})
}

func (h *Handler) getUserRatingHistory(c *gin.Context) {
	sid := c.Param("userId")
	userId, err := types.ParseObjectId(sid)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	p := &paginate.Paginated{
		Filters:     make(map[paginate.FilterParameter]paginate.Filter),
		Decscending: true,
	}

	ls, ok := c.GetQuery("limit")
	if ok {
		li, err := strconv.Atoi(ls)
		if err == nil {
			p.PerPage = uint64(li)
		}
	}

	ps, ok := c.GetQuery("page")
	if ok {
		pi, err := strconv.Atoi(ps)
		if err == nil {
			p.Page = uint64(pi)
		}
	}

	if err := p.Validate(); err != nil {
		c.JSON(400, gin.H{"error": "invalid pagination parameters"})
	}

	history, total, err := h.rating.GetUserChangeHistory(c, userId, p)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	res := UserRatingHistoryResponse{
		PaginatedResponseBase: paginate.PaginatedResponseBase{
			CurrentPage:  p.Page,
			PageSize:     uint64(len(history)),
			TotalNumbers: uint64(total),
			TotalPages:   (total + p.PerPage - 1) / p.PerPage,
		},
		List: make([]UserGameEloChange, 0, len(history)),
	}

	for _, change := range history {
		res.List = append(res.List, UserGameEloChange{
			UserId:     change.UserId.String(),
			GameId:     change.GameId.String(),
			OpponentId: change.OpponentId.String(),
			Change:     change.EloChange,
			Result:     change.Result.String(),
			Timestamp:  change.UpdatedAt.Unix(),
		})
	}
	c.JSON(200, res)
}
