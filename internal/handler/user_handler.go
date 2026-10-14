package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/common/response"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/middleware"
	"github.com/phong0206/go_gin_restful/internal/usecases"
)

type UserHandler struct {
	userUsecase *usecases.UserUsecase
}

func NewUserHandler(userUsecase *usecases.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

func (h *UserHandler) List(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"message": "User list endpoint",
		"users":   []any{},
	})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.RespondError(c, domainerrors.New(domainerrors.KindValidation, "INVALID_USER_ID", "user id must be a valid integer"))
		return
	}
	if id < 1 {
		middleware.RespondError(c, domainerrors.New(domainerrors.KindValidation, "INVALID_USER_ID", "user id must be a positive integer"))
		return
	}

	user, err := h.userUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, user)
}
