package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
)

type LessonHandler struct {
	LessonUseCase domain.LessonUseCase
	JWTUtil       *infra.JWTUtil
}

func NewLessonHandler(LessonUseCase domain.LessonUseCase, JWTUtil *infra.JWTUtil) *LessonHandler {
	handler := &LessonHandler{
		LessonUseCase: LessonUseCase,
		JWTUtil:       JWTUtil,
	}
	return handler
}

func (lh *LessonHandler) HandleGetLessonProgress(c echo.Context) (err error) {
	LessonUseCase := lh.LessonUseCase
	ju := lh.JWTUtil

	claims := ju.GetContextToken(c)
	user := new(domain.UserModel)
	user.ID = claims.UID

	progress, err := LessonUseCase.GetUserLessonProgress(c.Request().Context(), user)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, progress)
}
