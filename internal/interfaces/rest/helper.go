package rest

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

type endpoint struct {
	apiVersion  string
	middlewares []echo.MiddlewareFunc
	groups      []*apiGroup
}

type apiGroup struct {
	prefix      string
	middlewares []echo.MiddlewareFunc
	routes      []*route
}

type route struct {
	method      string
	path        string
	handler     echo.HandlerFunc
	middlewares []echo.MiddlewareFunc
}

func createEndpoint(app *echo.Echo, def *endpoint) {
	type RESTMethod func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

	var root *echo.Group
	if strings.HasPrefix(def.apiVersion, "/") {
		root = app.Group(def.apiVersion, def.middlewares...)
	} else {
		root = app.Group("/"+def.apiVersion, def.middlewares...)
	}

	for _, group := range def.groups {
		echoGroup := root.Group(group.prefix, group.middlewares...)
		for _, api := range group.routes {
			var method RESTMethod
			switch api.method {
			case "GET":
				method = echoGroup.GET
			case "POST":
				method = echoGroup.POST
			case "PUT":
				method = echoGroup.PUT
			case "DELETE":
				method = echoGroup.DELETE
			case "HEAD":
				method = echoGroup.HEAD
			case "CONNECT":
				method = echoGroup.CONNECT
			default:
				panic(fmt.Errorf("createEndpoint: unknown method %s", api.method))
			}
			method(api.path, api.handler, api.middlewares...)
		}
	}
}
