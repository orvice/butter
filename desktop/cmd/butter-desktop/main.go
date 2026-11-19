//go:build darwin

// Command butter-desktop is Butter's native macOS client.
package main

import (
	"log"

	"github.com/egoist/mygo"

	"go.orx.me/apps/butter/desktop/internal/app"
	"go.orx.me/apps/butter/desktop/internal/credentials"
	"go.orx.me/apps/butter/desktop/internal/prefs"
)

// secretService is the Keychain service the session token is stored under.
const secretService = "go.orx.me/apps/butter/desktop"

func main() {
	path, err := prefs.DefaultPath()
	if err != nil {
		log.Printf("desktop: preferences will not be saved: %v", err)
	}
	cfg, err := prefs.Load(path)
	if err != nil {
		log.Printf("desktop: ignoring broken preferences: %v", err)
	}

	a := app.New(app.Options{
		Prefs:     cfg,
		PrefsPath: path,
		Store:     credentials.NewSystem(secretService),
	})
	mygo.App.WhenReady(a.Start)
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
