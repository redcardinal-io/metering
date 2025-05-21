package planassignments

// func getPlanId(h *httpHandler, c context.Context, idOrSlug string) (*uuid.UUID, error) {
// 	// Try to parse as UUID first
// 	planId, err := uuid.Parse(idOrSlug)
// 	if err != nil {
// 		plan, getErr := h.planSvc.GetPlanByIDorSlug(c, idOrSlug)
// 		if getErr != nil {
// 			return nil, getErr
// 		}
// 		planId = plan.ID
// 	}
//
// 	return &planId, nil
// }
//
// func (h *httpHandler) assignPlan(ctx *fiber.Ctx) error {
// 	tenant_slug := ctx.Get(constants.TenantHeader)
// 	var req assignPlanRequest
//
// 	idOrSlug := ctx.Params("idOrSlug")
//
// 	if idOrSlug == "" {
// 		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "plan ID is required")
// 		h.logger.Error("plan idOrSlug is required", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
//
// 	if err := ctx.BodyParser(&req); err != nil {
// 		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "failed to parse request body")
// 		h.logger.Error("failed to parse request body", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
//
// 	// validate the request body
// 	if err := h.validator.Struct(req); err != nil {
// 		errResp := domainerrors.NewErrorResponseWithOpts(err, domainerrors.EINVALID, "invalid request body")
// 		h.logger.Error("invalid request body", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
//
// 	c := context.WithValue(ctx.UserContext(), constants.TenantSlugKey, tenant_slug)
//
// 	planId, getErr := getPlanId(h, c, idOrSlug)
//
// 	if getErr != nil {
// 		errResp := domainerrors.NewErrorResponseWithOpts(getErr, domainerrors.EINVALID, "invalid plan id or slug")
// 		h.logger.Error("invalid plan id or slug", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
//
// 	if req.OrganizationId != "" && req.UserId != "" {
// 		errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "organization_id and user_id are mutually exclusive, provide any one")
// 		h.logger.Error("organization_id and user_id are mutually exclusive, provide any one", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
//
// 	var timeFormaterr error
// 	var isOrg bool
// 	var orgOrUserId uuid.UUID
// 	var genErr error
// 	var planAssignment *models.PlanAssignment
//
// 	if req.OrganizationId != "" {
// 		isOrg = true
// 		orgOrUserId, genErr = uuid.Parse(req.OrganizationId)
// 		if genErr != nil {
// 			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse organization_id")
// 			h.logger.Error("unable to parse organization_id", zap.Reflect("error", errResp))
// 			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 		}
// 	} else {
// 		isOrg = false
// 		orgOrUserId, genErr = uuid.Parse(req.UserId)
// 		if genErr != nil {
// 			errResp := domainerrors.NewErrorResponseWithOpts(nil, domainerrors.EINVALID, "unable to parse user_id")
// 			h.logger.Error("unable to parse user_id", zap.Reflect("error", errResp))
// 			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 		}
// 	}
//
// 	ValidFrom, timeFormaterr := time.Parse(constants.TimeFormat, *req.ValidFrom)
// 	if timeFormaterr != nil {
// 		errResp := domainerrors.NewErrorResponseWithOpts(timeFormaterr, domainerrors.EINVALID, "invalid timestamp format")
// 		h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
// 		return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 	}
// 	if req.ValidUntil == nil {
// 		var ValidUntil pgtype.Timestamptz
//
// 		planAssignment, genErr = h.planSvc.AssignPlan(c, *planId, models.AssignOrUpdateAssignedPlanInput{
// 			OrganizationOrUserId: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
// 			ValidFrom:            pgtype.Timestamptz{Time: ValidFrom, Valid: true},
// 			ValidUntil:           ValidUntil,
// 			By:                   req.CreatedBy,
// 		}, isOrg)
//
// 		if genErr != nil {
// 			h.logger.Error("failed to assign plan", zap.Reflect("error", genErr))
// 			errResp := domainerrors.NewErrorResponse(genErr)
// 			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 		}
// 	} else {
// 		ValidUntil, timeFormaterr := time.Parse(constants.TimeFormat, *req.ValidUntil)
// 		if timeFormaterr != nil {
// 			errResp := domainerrors.NewErrorResponseWithOpts(timeFormaterr, domainerrors.EINVALID, "invalid timestamp format")
// 			h.logger.Error("invalid timestamp format", zap.Reflect("error", errResp))
// 			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 		}
//
// 		planAssignment, genErr = h.planSvc.AssignPlan(c, *planId, models.AssignOrUpdateAssignedPlanInput{
// 			OrganizationOrUserId: pgtype.UUID{Bytes: orgOrUserId, Valid: true},
// 			ValidFrom:            pgtype.Timestamptz{Time: ValidFrom, Valid: true},
// 			ValidUntil:           pgtype.Timestamptz{Time: ValidUntil, Valid: true},
// 			By:                   req.CreatedBy,
// 		}, isOrg)
//
// 		if genErr != nil {
// 			h.logger.Error("failed to assign plan", zap.Reflect("error", genErr))
// 			errResp := domainerrors.NewErrorResponse(genErr)
// 			return ctx.Status(errResp.Status).JSON(errResp.ToJson())
// 		}
// 	}
//
// 	return ctx.
// 		Status(fiber.StatusCreated).JSON(models.NewHttpResponse(planAssignment, "plan created successfully", fiber.StatusCreated))
// }
