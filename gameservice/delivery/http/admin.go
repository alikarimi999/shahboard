package http

func (r *Router) setupAdminRoutes() {
	r.gin.Group("/admin")

}
