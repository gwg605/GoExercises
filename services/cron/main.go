package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"valerygordeev/go/exercises/libs/base"
	"valerygordeev/go/exercises/libs/cron"

	"github.com/akamensky/argparse"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	_ "github.com/ncruces/go-sqlite3/vfs/memdb"
)

const (
	AppName = "cron"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	parser := argparse.NewParser("cronservice", "Cron Service")
	configFile := parser.String("c", "config", &argparse.Options{Required: true, Help: "Config file location"})
	useMachineConfig := parser.Flag("", "machine-config", &argparse.Options{Default: false, Help: "Use machine config path"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	log.Printf("main() - Initialization")

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = base.InitVars("", AppName, *useMachineConfig)
	if err != nil {
		log.Fatalf("Unable to init var folder. Error=%v", err)
	}
	log.Printf("main() - Vars inited: %v", base.GetGlobalValues())

	config, err := LoadCronConfig(*configFile)
	if err != nil {
		log.Fatalf("Unable to load config. Error=%v", err)
	}
	log.Printf("main() - Config loaded")

	db, err := base.OpenDatabase(config.DBConfig)
	if err != nil {
		log.Fatalf("Unable to open database. Error=%v", err)
	}
	log.Printf("main() - DB opened")

	service, err := cron.NewCronService(db, config.ServiceConfig)
	if err != nil {
		log.Fatalf("Unable to create service. Error=%v", err)
	}
	defer func() {
		service.Close()
		log.Printf("main() - Service stopped")
	}()
	log.Printf("main() - Service created")

	httpRoute := chi.NewRouter()
	httpRoute.Use(middleware.Logger)
	httpRoute.Mount("/cron/v1", cron.GetHttpServiceHandler(service))

	server, err := base.NewWebServer(&config.WebConfig, httpRoute)
	if err != nil {
		log.Fatalf("Unable to webserver. Error=%v", err)
	}
	log.Printf("main() - Web server created")

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		log.Printf("main() - Start listening on %s...", config.WebConfig.BindingAddress)
		err := server.ListenAndServer()
		log.Printf("main() - Listen stopped. Error=%v", err)
		return err
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Printf("main() - Shutdown webserver")
		_ = server.Shutdown(gCtx)
		return nil
	})
	err = g.Wait()
	log.Printf("main() - Group error=%v", err)
}
