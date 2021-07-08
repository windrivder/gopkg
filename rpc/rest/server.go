package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/windrivder/gopkg/errorx"
	"github.com/windrivder/gopkg/logx"
	"github.com/windrivder/gopkg/rpc/rest/middleware"
)

type Options struct {
	Name            string        `json:"name"`
	Mode            string        `json:"mode"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	CertFile        string        `json:"cert_file"`
	KeyFile         string        `json:"key_file"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	ClientTimeout   time.Duration `json:"client_timeout"`
	Secret          string        `json:"secret"`
	Expired         time.Duration `json:"expired"`
}

func NewOptions(v *viper.Viper) (o Options, err error) {
	if err = v.UnmarshalKey("rest", &o); err != nil {
		return o, errorx.Wrap(err, "unmarshal rest option error")
	}

	return o, err
}

type Server struct {
	*echo.Echo
	o   Options
	log logx.Logger
}

func New(o Options, log logx.Logger, fn HandlerRoutersFunc) (IServer, func(), error) {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	s := &Server{o: o, log: log, Echo: e}

	fn(s)

	return s, func() { s.Stop() }, nil
}

func (s *Server) Start() (err error) {
	addr := fmt.Sprintf("%s:%d", s.o.Host, s.o.Port)
	s.log.WithFields(logx.Fields{"addr": addr}).Info("http server starting...")

	go func() {
		if s.o.CertFile == "" && s.o.KeyFile == "" {
			err = s.Echo.Server.ListenAndServe()
		} else {
			err = s.Echo.Server.ListenAndServeTLS(s.o.CertFile, s.o.KeyFile)
		}

		if err != nil && err != http.ErrServerClosed {
			s.log.Fatalf("start http server err: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.log.Info("http server stopping...")

	timeout := time.Second * s.o.ShutdownTimeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.Echo.Shutdown(ctx)
}

var ProviderSet = wire.NewSet(New, NewOptions)