package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gotk/ctx"
	"github.com/gotk/pg"

	"github.com/vonwenm/go-spa/backend/account"
	"github.com/vonwenm/go-spa/backend/account/group"
	"github.com/vonwenm/go-spa/backend/account/resetpassword"
	"github.com/vonwenm/go-spa/backend/account/token"
	"github.com/vonwenm/go-spa/backend/account/user"
	"github.com/vonwenm/go-spa/backend/cfg"
	"github.com/vonwenm/go-spa/backend/storage/location"
	"github.com/vonwenm/go-spa/backend/storage/media"
	"github.com/vonwenm/go-spa/backend/storage/mediatype"
	"github.com/vonwenm/go-spa/backend/storage/mediaupload"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func main() {
	cfg.MustLoad()
	listeningAddr := ":" + cfg.Server.Port

	router := mux.NewRouter().PathPrefix(cfg.Server.API.Prefix).Subrouter()

	db, err := pg.NewSession(cfg.DB.ConnectionURL())
	if err != nil {
		log.Fatalf("Unable to connect to database: %s", err)
	}

	vars := map[string]interface{}{"db": db}

	err = ctx.Init(router, cfg.Server.API.PrivKey, cfg.Server.API.PubKey, vars)
	if err != nil {
		log.Fatalf("Unable to configure API: %s", err)
	}

	server := negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(WebLogger),
		negroni.NewStatic(http.Dir(cfg.Server.Frontend.Path)),
	)
	server.UseHandler(router)

	log.Infof("Listening on address: %s", listeningAddr)
	log.Fatal(http.ListenAndServe(listeningAddr, server))
}

func WebLogger(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, req)
	res := rw.(negroni.ResponseWriter)
	defer func() {
		elapsed := time.Since(start)
		log.WithFields(log.Fields{
			"elapsed": elapsed,
			"method":  req.Method,
			"host":    req.URL.Host,
			"path":    req.URL.Path,
			"query":   req.URL.RawQuery,
			"status":  res.Status(),
			"size":    res.Size(),
		}).Info(req.Method + " " + req.URL.Path)
	}()
}
