package cmd

import (
	"time"

	"github.com/RichardKnop/example-api/services/accounts"
	"github.com/RichardKnop/example-api/services/email"
	"github.com/RichardKnop/example-api/services/facebook"
	"github.com/RichardKnop/example-api/services/health"
	"github.com/RichardKnop/example-api/services/oauth"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"
)

// RunServer runs the app
func RunServer() error {
	cnf, db, err := initConfigDB(true, true)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialise the health service
	healthService := health.NewService(db)

	// Initialise the oauth service
	oauthService := oauth.NewService(cnf, db)

	// Initialise the email service
	emailService := email.NewService(cnf)

	// Initialise the accounts service
	accountsService := accounts.NewService(
		cnf,
		db,
		oauthService,
		emailService,
		nil, // accounts.EmailFactory
	)

	// Initialise the facebook service
	facebookService := facebook.NewService(
		cnf,
		db,
		accountsService,
		nil, // facebook.Adapter
	)

	// Start a negroni app
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(gzip.Gzip(gzip.DefaultCompression))

	// Create a router instance
	router := mux.NewRouter()

	// Register routes
	healthService.RegisterRoutes(router, "/v1")
	oauthService.RegisterRoutes(router, "/v1/oauth")
	accountsService.RegisterRoutes(router, "/v1")
	facebookService.RegisterRoutes(router, "/v1/facebook")

	// Set the router
	app.UseHandler(router)

	// Run the server on port 8080, gracefully stop on SIGTERM signal
	graceful.Run(":8080", 5*time.Second, app)

	return nil
}
