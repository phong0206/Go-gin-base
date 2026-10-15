package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phong0206/go_gin_restful/internal/common/response"
	domainerrors "github.com/phong0206/go_gin_restful/internal/domain/errors"
	"github.com/phong0206/go_gin_restful/internal/dto"
	"github.com/phong0206/go_gin_restful/internal/middleware"
	"github.com/phong0206/go_gin_restful/internal/usecases"
)

type AuthHandler struct {
	authUsecase *usecases.AuthUsecase
}

func NewAuthHandler(authUsecase *usecases.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, domainerrors.New(
			domainerrors.KindValidation,
			"INVALID_INPUT",
			err.Error(),
		))
		return
	}

	result, err := h.authUsecase.Login(c.Request.Context(), req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, domainerrors.New(
			domainerrors.KindValidation,
			"INVALID_INPUT",
			err.Error(),
		))
		return
	}

	result, err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, result)
}
