// Command butter-delete-web-chat deletes the data the old dashboard Chat
// left behind (#411). Chat runs on AG-UI since #409, and the old Chat's
// "web-chat" sessions are not kept: no migration and no read-only view.
//
// It is a dry run unless --confirm is given. Never run it against production
// without the project owner's explicit confirmation; see usageText.
//
// The cleanup itself is internal/maintenance/webchatcleanup, which the
// service also runs on every startup for now (#411).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gopkg.in/yaml.v3"

	appconfig "go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
)

const commandName = "butter-delete-web-chat"

// envConfigPath names Butter's config file, as the service reads it.
const envConfigPath = "BUTTERFLY_CONFIG_FILE_PATH"

// The service's own defaults when its config names no MongoDB.
const (
	defaultMongoURI = "mongodb://localhost:27017"
	defaultMongoDB  = "butter"
)

const usageText = `Usage: butter-delete-web-chat [--config FILE] [--mongo-uri URI] [--mongo-db NAME] [--confirm]

Deletes the data the old dashboard Chat left behind (#411), in this order:
  1. invocation_input_parts of the Invocations below;
  2. invocations whose session_id is one of the web-chat sessions', except
     any whose record names another app;
  3. adk_events of the web-chat sessions;
  4. adk_sessions with app_name "web-chat".

!!! WARNING !!!
Never run it against production without the project owner's explicit
confirmation, given at the time, after they reviewed a dry run, and after a
backup or snapshot of the database. Run it only once #409 is deployed, so no
new web-chat sessions appear while it runs.

Without --confirm it is a dry run: it prints the counts per collection and per
workspace, and changes nothing. With --confirm it deletes them. It is
idempotent: running it again deletes nothing more and reports zero.

Deleting the Invocations lowers past Activity counts on the dashboard, which
counts Invocations by time range. Workspace Memory (mem0) entries and every
other app's data are not touched.

It reads mongo_uri and mongo_db from Butter's config file, as the service does
(default: mongodb://localhost:27017 and butter). --mongo-uri and --mongo-db
override them.

Flags:
`

func main() {
	os.Exit(cli(os.Args[1:], os.Stdout, os.Stderr))
}

// cli runs the command and returns its exit code: 0 on success, 1 when the
// cleanup fails, 2 on bad usage.
func cli(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "Butter config file to read mongo_uri and mongo_db from (default $"+envConfigPath+")")
	mongoURI := fs.String("mongo-uri", "", "MongoDB URI; overrides the config file's mongo_uri")
	mongoDB := fs.String("mongo-db", "", "MongoDB database; overrides the config file's mongo_db")
	confirm := fs.Bool("confirm", false, "delete the data; without it, only report what would be deleted")
	fs.Usage = func() {
		_, _ = fmt.Fprint(stderr, usageText)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "%s: unexpected arguments: %s\n", commandName, strings.Join(fs.Args(), " "))
		fs.Usage()
		return 2
	}

	t, err := resolveTarget(*configPath, *mongoURI, *mongoDB)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%s: %v\n", commandName, err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := cleanup(ctx, t, *confirm, stdout); err != nil {
		_, _ = fmt.Fprintf(stderr, "%s: %v\n", commandName, err)
		return 1
	}
	return 0
}

// target is the database the cleanup works on.
type target struct {
	uri string
	db  string
}

// resolveTarget reads mongo_uri and mongo_db as the service does: from the
// YAML config file at configPath, else at $BUTTERFLY_CONFIG_FILE_PATH, with
// the service's defaults for what it leaves out. Non-empty flag values win.
func resolveTarget(configPath, uriFlag, dbFlag string) (target, error) {
	t := target{uri: defaultMongoURI, db: defaultMongoDB}
	if configPath == "" {
		configPath = os.Getenv(envConfigPath)
	}
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return target{}, fmt.Errorf("read config: %w", err)
		}
		var cfg appconfig.AppConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return target{}, fmt.Errorf("parse config %s: %w", configPath, err)
		}
		if cfg.MongoURI != "" {
			t.uri = cfg.MongoURI
		}
		if cfg.MongoDB != "" {
			t.db = cfg.MongoDB
		}
	}
	if uriFlag != "" {
		t.uri = uriFlag
	}
	if dbFlag != "" {
		t.db = dbFlag
	}
	return t, nil
}

// cleanup connects to the target and runs the cleanup on it. It names the
// database and its hosts first, never the URI, which may hold credentials.
func cleanup(ctx context.Context, t target, confirm bool, out io.Writer) error {
	opts := options.Client().ApplyURI(t.uri)
	client, err := mongo.Connect(opts)
	if err != nil {
		return fmt.Errorf("connect to mongodb: %w", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("reach mongodb at %s: %w", strings.Join(opts.Hosts, ","), err)
	}

	mode := "DRY RUN"
	if confirm {
		mode = "DELETING (--confirm)"
	}
	_, _ = fmt.Fprintf(out, "%s: %s on database %q at %s\n\n", commandName, mode, t.db, strings.Join(opts.Hosts, ","))
	_, err = webchatcleanup.Run(ctx, client.Database(t.db), confirm, out)
	return err
}
