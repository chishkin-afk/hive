package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/gin-gonic/gin"
)

type HandlerFuncs struct {
	authService     authService
	offerService    offerService
	clientService   clientService
	scheduleService scheduleService
	bookingService  bookingService
}

func (h *HandlerFuncs) parsePagination(ctx *gin.Context) (uint32, uint32, error) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 1000 {
		return 0, 0, errs.ErrInvalidPage
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, errs.ErrInvalidLimit
	}

	return uint32(page), uint32(limit), nil
}

func (h *HandlerFuncs) getCodeFromError(err error) int {
	var errKind *errs.ErrorKind
	if !errors.As(err, &errKind) {
		return http.StatusInternalServerError
	}

	switch errKind.Kind() {
	case errs.KindRequest:
		return http.StatusBadRequest
	case errs.KindInternal:
		return http.StatusInternalServerError
	case errs.KindUnauthorized:
		return http.StatusUnauthorized
	case errs.KindConflict:
		return http.StatusConflict
	case errs.KindTimeout:
		return http.StatusRequestTimeout
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindPermissionDenied:
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}
