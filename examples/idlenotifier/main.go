package main

import (
	"fmt"
	"github.com/vcaesar/go-wayland/client"
	idlenotify "github.com/vcaesar/go-wayland/staging/ext-idle-notify-v1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Global app state
type appState struct {
	display  *client.Display
	registry *client.Registry
	seat     *client.Seat
	notifier *idlenotify.IdleNotifier
}

func main() {
	app := &appState{}

	app.initWindow()

	dispatchQueue := make(chan func() error)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			dispatchFunc := app.display.Context().GetDispatch()
			dispatchQueue <- dispatchFunc
		}
	}()

	stop := time.After(15 * time.Second)

	for {
		select {
		// Code that does other things can be added here.
		case <-stop:
			app.cleanup()
			return
		case <-c:
			app.cleanup()
			os.Exit(1)
			return
		case dispatchFunc := <-dispatchQueue:
			// It is important to call dispatchFunc in the same goroutine.
			err := dispatchFunc()
			if err != nil {
				log.Printf("Error dispatching: %v\n", err)
			}
		}
	}
}

func (app *appState) initWindow() {
	// Connect to wayland server
	display, err := client.Connect("")
	if err != nil {
		log.Fatalf("unable to connect to wayland server: %v", err)
	}
	app.display = display

	display.SetErrorHandler(app.HandleDisplayError)

	// Get global interfaces registry
	registry, err := app.display.GetRegistry()
	if err != nil {
		log.Fatalf("unable to get global registry object: %v", err)
	}
	app.registry = registry

	// Add global interfaces registrar handler
	registry.SetGlobalHandler(app.HandleRegistryGlobal)
	// Wait for interfaces to register
	app.display.Roundtrip()
	// Wait for handler events
	app.display.Roundtrip()

	notification, err := app.notifier.GetIdleNotification(5000, app.seat)
	if err != nil {
		log.Fatalf("Error getting idle notification: %v", err)
	}
	notification.SetIdledHandler(func(event idlenotify.IdleNotificationIdledEvent) {
		fmt.Printf("Idle\n")
	})
	notification.SetResumedHandler(func(event idlenotify.IdleNotificationResumedEvent) {
		fmt.Printf("Resumed\n")
	})
}

func (app *appState) context() *client.Context {
	return app.display.Context()
}

func (app *appState) HandleRegistryGlobal(e client.RegistryGlobalEvent) {
	switch e.Interface {
	case idlenotify.IdleNotifierInterfaceName:
		app.notifier = idlenotify.NewIdleNotifier(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, app.notifier)
		if err != nil {
			log.Fatalf("Unable to bind %s interface: %v", idlenotify.IdleNotifierInterfaceName, err)
		}
	case client.SeatInterfaceName:
		seat := client.NewSeat(app.context())
		err := app.registry.Bind(e.Name, e.Interface, e.Version, seat)
		if err != nil {
			log.Fatalf("unable to bind %s interface: %v", client.SeatInterfaceName, err)
		}
		app.seat = seat
	}
}

// HandleDisplayError handles client.Display errors
func (*appState) HandleDisplayError(e client.DisplayErrorEvent) {
	// Just log.Fatal for now
	log.Fatalf("display error event: %v", e)
}

func (app *appState) cleanup() {
	// Release wl_seat handlers
	if app.seat != nil {
		if err := app.seat.Release(); err != nil {
			log.Println("unable to release wl_seat:", err)
		}
		app.seat = nil
	}

	if app.registry != nil {
		if err := app.registry.Destroy(); err != nil {
			log.Println("unable to destroy wl_registry:", err)
		}
		app.registry = nil
	}

	if app.display != nil {
		if err := app.display.Destroy(); err != nil {
			log.Println("unable to destroy wl_display:", err)
		}
	}

	if app.notifier != nil {
		if err := app.notifier.Destroy(); err != nil {
			log.Printf("unable to destroy %s, %v:\n", idlenotify.IdleNotifierInterfaceName, err)
		}
	}

	// Close the wayland server connection
	if err := app.context().Close(); err != nil {
		log.Println("unable to close wayland context:", err)
	}
}
