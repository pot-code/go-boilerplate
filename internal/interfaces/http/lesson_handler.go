package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/lesson"
	"github.com/pot-code/go-boilerplate/internal/user"
)

type LessonHandler struct {
	lessonUseCase lesson.LessonUseCase
	jwtUtil       *auth.JWTUtil
}

func NewLessonHandler(LessonUseCase lesson.LessonUseCase, JWTUtil *auth.JWTUtil) *LessonHandler {
	handler := &LessonHandler{LessonUseCase, JWTUtil}
	return handler
}

func (lh *LessonHandler) HandleGetLessonProgress(c echo.Context) (err error) {
	LessonUseCase := lh.lessonUseCase
	ju := lh.jwtUtil

	claims := ju.GetContextToken(c)
	user := new(user.UserModel)
	user.ID = claims.UID

	progress, err := LessonUseCase.GetUserLessonProgress(c.Request().Context(), user)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, progress)
}
