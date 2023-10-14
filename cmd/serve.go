package cmd

import (
	"github.com/crosstyan/dumb_downloader/global/log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"moul.io/chizap"

	// swagger embed files
	_ "github.com/crosstyan/dumb_downloader/docs"
)

// @title Dumb Downloader API
// @version 1.0
// @license.name Do What the Fuck You Want to Public License
// @license.url http://www.wtfpl.net/
func serveRun(cmd *cobra.Command, args []string) {
	listenAddr, err := GetListenAddrFromViper()
	if err != nil {
		log.Sugar().Panicw("failed to get listen address", "error", err)
	}
	r := chi.NewRouter()
	// middleware
	chiZapM := chizap.New(log.Logger(), &chizap.Opts{})
	corsM := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	swaggerH := httpSwagger.Handler(
		// The url pointing to API definition
		httpSwagger.URL("/swagger/doc.json"),
	)
	r.Use(chiZapM, corsM)
	r.Get("/swagger/*", swaggerH)
	err = http.ListenAndServe(listenAddr, r)
	if err != nil {
		log.Sugar().Panicw("listen", "err", err)
	}
}

func tryDownload() {

}

var serve = cobra.Command{
	Use:   "serve",
	Short: "serve a dumb downloader server",
	Args:  cobra.NoArgs,
	Run:   serveRun,
}
