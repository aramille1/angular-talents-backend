// Add these methods to your existing RecruiterController struct

// GetRecruiters retrieves recruiters with optional filtering by status
// @Summary Get recruiters
// @Description Get recruiters with optional filtering by status
// @Tags admin,recruiters
// @Produce json
// @Param status query string false "Filter by status (pending, approved, rejected)"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/recruiters [get]
func (c *RecruiterController) GetRecruiters(ctx *gin.Context) {
	// Parse pagination params
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	status := ctx.Query("status")

	// Get recruiters
	recruiters, total, err := c.recruiterService.GetRecruiters(ctx.Request.Context(), page, limit, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"recruiters": recruiters,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

// GetPendingRecruiters retrieves pending recruiters
// @Summary Get pending recruiters
// @Description Get recruiters with pending status
// @Tags admin,recruiters
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.RecruiterResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/recruiters/pending [get]
func (c *RecruiterController) GetPendingRecruiters(ctx *gin.Context) {
	// Get pending recruiters
	recruiters, err := c.recruiterService.GetRecruitersByStatus(ctx.Request.Context(), "pending")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, recruiters)
}

// ApproveRecruiter approves a recruiter
// @Summary Approve recruiter
// @Description Approve a recruiter, changing their status to approved
// @Tags admin,recruiters
// @Produce json
// @Param id path string true "Recruiter ID"
// @Security ApiKeyAuth
// @Success 200 {object} models.RecruiterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/recruiters/{id}/approve [patch]
func (c *RecruiterController) ApproveRecruiter(ctx *gin.Context) {
	id := ctx.Param("id")

	// Get admin ID from context
	adminInterface, exists := ctx.Get("admin")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Admin info not found"})
		return
	}

	admin, ok := adminInterface.(*models.AdminResponse)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin info"})
		return
	}

	// Approve recruiter
	updatedRecruiter, err := c.recruiterService.UpdateRecruiterStatus(
		ctx.Request.Context(),
		id,
		"approved",
		admin.ID.Hex(),
		"",
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecruiter)
}

// RejectRecruiter rejects a recruiter
// @Summary Reject recruiter
// @Description Reject a recruiter, changing their status to rejected
// @Tags admin,recruiters
// @Accept json
// @Produce json
// @Param id path string true "Recruiter ID"
// @Param reason body map[string]string false "Rejection reason"
// @Security ApiKeyAuth
// @Success 200 {object} models.RecruiterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/recruiters/{id}/reject [patch]
func (c *RecruiterController) RejectRecruiter(ctx *gin.Context) {
	id := ctx.Param("id")

	// Get rejection reason from request body
	var reqBody struct {
		Reason string `json:"reason"`
	}

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		// If there's no body or invalid JSON, continue without a reason
		reqBody.Reason = ""
	}

	// Get admin ID from context
	adminInterface, exists := ctx.Get("admin")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Admin info not found"})
		return
	}

	admin, ok := adminInterface.(*models.AdminResponse)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin info"})
		return
	}

	// Reject recruiter
	updatedRecruiter, err := c.recruiterService.UpdateRecruiterStatus(
		ctx.Request.Context(),
		id,
		"rejected",
		admin.ID.Hex(),
		reqBody.Reason,
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecruiter)
}
