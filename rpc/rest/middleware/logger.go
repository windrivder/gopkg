package middleware

import (
	"strings"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/windrivder/gopkg/logx"
)

type (
	LoggerConfig struct {
		Skipper SkipperFunc
		Trans   ut.Translator
	}
)

var (
	DefaultLoggerConfig = LoggerConfig{
		Skipper: DefaultSkipper,
	}
)

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			path := req.URL.Path
			if path == "" {
				path = "/"
			}

			var log *logx.Event
			if err != nil {
				log = logx.Error()

				// 打印请求校验信息
				rerr, ok := err.(validator.ValidationErrors)
				if ok {
					if config.Trans != nil {
						for field, msg := range rerr.Translate(config.Trans) {
							log = log.Str(field[strings.Index(field, ".")+1:], msg)
						}
					} else {
						log = log.Err(rerr)
					}
				} else {
					log = logx.Err(err)
				}
			} else {
				log = logx.Info()
			}

			stop := time.Now()
			log.Str("id", id).
				Str("path", path).
				Str("remote_ip", c.RealIP()).
				Str("uri", req.RequestURI).
				Str("method", req.Method).
				Int("status", res.Status).
				Str("latency", stop.Sub(start).String()).Send()

			return nil
		}
	}
}
