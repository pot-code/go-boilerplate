package http

import (
	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
)

func v1Endpoint(
	websocket *infra.Websocket,
	UserHandler *UserHandler,
	LessonHandler *LessonHandler,
	TimeSpentHandler *TimeSpentHandler,
	jwtMiddleware echo.MiddlewareFunc,
	refreshMiddleware echo.MiddlewareFunc,
	requestIDMiddleware echo.MiddlewareFunc,
	traceLoggerMiddleware echo.MiddlewareFunc,
) *endpoint {
	return &endpoint{
		apiVersion:  "api/v1",
		middlewares: []echo.MiddlewareFunc{requestIDMiddleware, traceLoggerMiddleware},
		groups: []*apiGroup{
			{
				prefix: "/user",
				routes: []*route{
					{"POST", "/login", UserHandler.HandleSignIn, nil},
					{"PUT", "/sign-out", UserHandler.HandleSignOut, nil},
					{"POST", "/sign-up", UserHandler.HandleSignUp, nil},
					{"GET", "/exists", UserHandler.HandleUserExists, nil},
				},
			},
			{
				prefix:      "/lesson",
				middlewares: []echo.MiddlewareFunc{jwtMiddleware, refreshMiddleware},
				routes: []*route{
					{"GET", "/progress", LessonHandler.HandleGetLessonProgress, nil},
				},
			},
			{
				prefix:      "/time-spent",
				middlewares: []echo.MiddlewareFunc{jwtMiddleware, refreshMiddleware},
				routes: []*route{
					{"GET", "/", TimeSpentHandler.HandleGetTimeSpent, nil},
				},
			},
			{
				prefix: "/ws",
				routes: []*route{
					{"GET", "/echo", websocket.WithHeartbeat(HandleEcho), nil},
				},
			},
		},
	}
}
