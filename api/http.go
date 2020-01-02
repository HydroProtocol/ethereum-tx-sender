package api

import (
	"context"
	"encoding/json"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/messages"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

func StartHTTPServer(ctx context.Context) {
	logrus.SetLevel(logrus.DebugLevel)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		MaxAge:       86400,
	}))
	loadRoutes(e)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%s", "3001"),
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	go func() {
		if err := e.StartServer(s); err != nil {
			e.Logger.Info("shutting down the server: %v", err)
			panic(err)
		}
	}()

	<-ctx.Done()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func CreateLog(p Param) (interface{}, error) {
	params := p.(*CreateLogReq)
	return createLog(&params.CreateMessage)
}

func GetLogs(p Param) (interface{}, error) {
	params := p.(*GetLogReq)
	return getLog(&params.GetMessage)
}

func loadRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	addRoute(e, "create_log", http.MethodPost, "/logs", &CreateLogReq{}, CreateLog)
	addRoute(e, "get_log", http.MethodGet, "/logs", &GetLogReq{}, GetLogs)
}

func addRoute(
		e *echo.Echo, apiKey, method, url string,
		param Param,
		handler func(p Param) (interface{}, error),
		middleWares ...echo.MiddlewareFunc) {

	routeHandlerFunc := commonHandler(apiKey, param, handler)
	e.Add(method, url, routeHandlerFunc, middleWares...)
}

func bindAndValidParams(c echo.Context, params Param) (err error) {
	json.NewDecoder(c.Request().Body).Decode(params)
	return nil
}

func commonHandler(apiKey string, paramSchema Param, fn func(Param) (interface{}, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		var err error
		startTime := time.Now()
		var resp interface{}
		var returnContent string
		var req Param
		defer func() {
			var ok bool
			if r := recover(); r != nil {
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				utils.Errorf("unhandled error: %v %s", err, string(stack[:length]))
			}

			status := "success"
			if err != nil {
				status = "failed"
			}

			cost := float64(time.Since(startTime) / 1000000)
			utils.Infof("######[%s] [%.0f] [%s]######request[%s]######response[%s]######err[%v]", apiKey, cost, status, utils.ToJsonString(req), returnContent, err)
		}()

		if paramSchema != nil {
			// Params is shared among all request
			// We create a new one for each request
			req = reflect.New(reflect.TypeOf(paramSchema).Elem()).Interface().(Param)
			err = bindAndValidParams(c, req)
		}

		if err != nil {
			returnContent = utils.ToJsonString(BaseResp{
				Status: -2,
				Desc:   err.Error(),
				Data:   resp,
			})
		} else {
			resp, err = fn(req)
			if err == nil {
				returnContent = utils.ToJsonString(BaseResp{
					Status: 0,
					Desc:   "success",
					Data:   resp,
				})
			} else {
				returnContent = utils.ToJsonString(BaseResp{
					Status:   -1,
					Desc:     err.Error(),
					Data:     resp,
				})
			}
		}

		_ = c.String(http.StatusOK, returnContent)
		return nil
	}
}

type (
	BaseReq struct {
	}

	BaseResp struct {
		Status   int               `json:"status"`
		Desc     interface{}       `json:"desc"`
		Data     interface{}       `json:"data,omitempty"`
	}

	CreateLogReq struct {
		BaseReq
		messages.CreateMessage
	}

	CreateLogResp struct {
		messages.CreateReply
	}

	GetLogReq struct {
		BaseReq
		messages.GetMessage
	}

	GetLogResp struct {
		messages.GetReply
	}
)

func (b *BaseReq) GetTrace() string {
	return ""
}

func (b *BaseReq) SetTrace(address string) {
}

type Param interface {
	GetTrace() string
	SetTrace(address string)
}

