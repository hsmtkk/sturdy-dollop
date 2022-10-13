package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firestoregollira "github.com/GoogleCloudPlatform/firestore-gorilla-sessions"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	portStr, err := requiredEnvVar("PORT")
	if err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err)
	}

	projectID, err := requiredEnvVar("GOOGLE_CLOUD_PROJECT")
	if err != nil {
		log.Fatal(err)
	}

	hdl, err := newHandler(projectID)
	if err != nil {
		log.Fatal(err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hdl.index)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

type myHandler struct {
	store sessions.Store
	tmpl  *template.Template
}

func newHandler(projectID string) (*myHandler, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore.NewClient failed; %w", err)
	}
	store, err := firestoregollira.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("firestoregollira.New failed; %w", err)
	}

	tmpl, err := template.New("Index").Parse(`<body>{{.views}} {{if eq .views 1.0}}view{{else}}views{{end}} for "{{.greeting}}"</body>`)
	if err != nil {
		return nil, fmt.Errorf("template.New.Parse failed; %w", err)
	}

	return &myHandler{store, tmpl}, nil
}

var greetings = []string{
	"Hello World",
	"Hallo Welt",
	"Ciao Mondo",
	"Salut le Monde",
	"Hola Mundo",
}

// Handler
func (h *myHandler) index(c echo.Context) error {
	name := "hello-views"
	session, err := h.store.Get(c.Request(), name)
	if err != nil {
		return fmt.Errorf("store.Get failed; %w", err)
	}
	if session.IsNew {
		session.Values["views"] = float64(0)
		session.Values["greeting"] = greetings[rand.Intn(len(greetings))]
	}
	session.Values["views"] = session.Values["views"].(float64) + 1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return fmt.Errorf("session.Save failed; %w", err)
	}
	if err := h.tmpl.Execute(c.Response(), session.Values); err != nil {
		return fmt.Errorf("tmplate.Execute failed; %w", err)
	}
	return c.NoContent(http.StatusOK)
}

func requiredEnvVar(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("you must define %s env var", key)
	}
	return val, nil
}
