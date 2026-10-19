// main.go
// Main SSH server for the game
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/charmbracelet/wish/recover"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/batt0s/tamagotchi.ssh/internal/session"
	"github.com/batt0s/tamagotchi.ssh/internal/storage"
	"github.com/batt0s/tamagotchi.ssh/internal/tui"
)

const (
	addr        = ":2222"
	hostKeyPath = ".ssh/pet_ed25519"
	dataDir     = "data"
	balancePath = "balance.yaml"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	balance, err := game.LoadBalance(balancePath)
	if err != nil {
		slog.Error("couldn't load balance", "err", err)
		os.Exit(1)
	}

	petStorage, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		slog.Error("couldn't load storage", "err", err)
		os.Exit(1)
	}

	sessions := session.NewManager()

	handler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		return teaHandler(s, balance, petStorage, sessions)
	}

	limiter := ratelimiter.NewRateLimiter(rate.Limit(5), 10, 128)

	srv, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithIdleTimeout(5*time.Minute),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			recover.Middleware(
				ratelimiter.Middleware(limiter),
				logging.Middleware(),
				bubbletea.Middleware(handler),
			),
		),
	)

	if err != nil {
		slog.Error("Couldn't create the server", "err", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("The server is listening on", "addr", addr, "host_key", hostKeyPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			slog.Error("Error while listening", "err", err)
			done <- syscall.SIGINT
		}
	}()

	<-done
	slog.Info("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func teaHandler(
	s ssh.Session,
	balance *game.Balance,
	petStorage storage.Storage,
	sessions *session.Manager,
) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	renderer := bubbletea.MakeRenderer(s)

	var fp string

	if pk := s.PublicKey(); pk != nil {
		fp = gossh.FingerprintSHA256(pk)
	} else {
		return tui.NewMessage("Couldn't verify Identity", pty.Window.Width, pty.Window.Height, renderer), nil
	}

	if !sessions.TryAcquire(fp) {
		slog.Info("Double Session Denied", "fp", fp)
		return tui.NewMessage(
			"You are already connected. Double sessions are not allowed.",
			pty.Window.Width, pty.Window.Height, renderer,
		), nil
	}

	now := time.Now()
	p, err := petStorage.LoadOrCreate(fp, now)
	if err != nil {
		sessions.Release(fp)
		slog.Error("couldn't load pet", "fp", fp, "err", err)
		return tui.NewMessage("Couldn't load pet. \nError:"+err.Error(), pty.Window.Width, pty.Window.Height, renderer), nil
	}

	offlineDuration := time.Duration(0)
	if !p.LastCalc.IsZero() && p.LastCalc.Before(now) {
		offlineDuration = now.Sub(p.LastCalc)
	}

	p.Apply(now, balance)

	go func() {
		<-s.Context().Done()
		sessions.Release(fp)
		slog.Info("logout", "fp", fp)
	}()

	slog.Info("login", "fp", fp,
		"satiety", p.Satiety, "happiness", p.Happiness, "energy", p.Energy)

	return tui.New(tui.Deps{
		Fingerprint: fp,
		Pet:         p,
		Balance:     balance,
		Storage:     petStorage,
		OfflineTime: offlineDuration,
		Renderer:    renderer,
	}, pty.Window.Width, pty.Window.Height), []tea.ProgramOption{tea.WithAltScreen()}
}
