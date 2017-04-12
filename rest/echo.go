package rest

import (
	"bytes"
	"strconv"
	"net/http"
	"io/ioutil"
	"github.com/labstack/echo"
	"github.com/buduchail/calavera"
)

type (
	EchoAPI struct {
		e      *echo.Echo
		prefix string
	}
)

func NewEcho(prefix string) (api EchoAPI) {
	api = EchoAPI{}
	api.e = echo.New()
	api.prefix = normalizePrefix(prefix)
	return api
}

func (api EchoAPI) getBody(c echo.Context) calavera.Payload {
	b, _ := ioutil.ReadAll(c.Request().Body)
	return bytes.NewBuffer(b).Bytes()
}

func (api EchoAPI) getQueryParameters(c echo.Context) calavera.QueryParameters {
	return calavera.QueryParameters(c.QueryParams())
}

func (api EchoAPI) getParentIds(c echo.Context, idParams []string) (ids []calavera.ResourceID) {
	ids = make([]calavera.ResourceID, 0)
	for _, id := range idParams {
		// prepend: /grandparent/1/parent/2/child/3 -> [2,1]
		ids = append([]calavera.ResourceID{calavera.ResourceID(c.Param(id))}, ids...)
	}
	return ids
}

func (api EchoAPI) sendResponse(c echo.Context, code int, body calavera.Payload, err error) error {

	if code != http.StatusOK || err != nil {
		if err == nil {
			err = getHttpError(code)
		}
		return c.String(code, err.Error())
	}

	return c.String(code, string(body))
}

func (api EchoAPI) AddResource(name string, handler calavera.ResourceHandler) {

	path, parentIdParams, idParam := expandPath(name, ":%s")

	postRoute := func(c echo.Context) error {
		code, body, err := handler.Post(
			api.getParentIds(c, parentIdParams),
			api.getBody(c),
		)
		return api.sendResponse(c, code, body, err)
	}

	getRoute := func(c echo.Context) error {
		code, body, err := handler.Get(
			calavera.ResourceID(c.Param(idParam)),
			api.getParentIds(c, parentIdParams),
		)
		return api.sendResponse(c, code, body, err)
	}

	getManyRoute := func(c echo.Context) error {
		code, body, err := handler.GetMany(
			api.getParentIds(c, parentIdParams),
			api.getQueryParameters(c),
		)
		return api.sendResponse(c, code, body, err)
	}

	putRoute := func(c echo.Context) error {
		code, body, err := handler.Put(
			calavera.ResourceID(c.Param(idParam)),
			api.getParentIds(c, parentIdParams),
			api.getBody(c),
		)
		return api.sendResponse(c, code, body, err)
	}

	deleteRoute := func(c echo.Context) error {
		code, body, err := handler.Delete(
			calavera.ResourceID(c.Param(idParam)),
			api.getParentIds(c, parentIdParams),
		)
		return api.sendResponse(c, code, body, err)
	}

	fullPath := api.prefix + path

	api.e.POST(fullPath, postRoute)
	api.e.POST(fullPath+"/", postRoute)

	api.e.GET(fullPath+"/:"+idParam, getRoute)
	api.e.GET(fullPath, getManyRoute)
	api.e.GET(fullPath+"/", getManyRoute)

	api.e.PUT(fullPath+"/:"+idParam, putRoute)

	api.e.DELETE(fullPath+"/:"+idParam, deleteRoute)
}

func (api EchoAPI) AddMiddleware(m calavera.Middleware) {
	// NOT IMPLEMENTED
}

func (api EchoAPI) Run(port int) {
	api.e.Start(":" + strconv.Itoa(port))
}
