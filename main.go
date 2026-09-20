package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/colorprofile"
)

const (
	host = "localhost"
	port = "23234"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(programHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("starting SSH server on %s:%s", host, port)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Printf("could not start server: %v", err)
			done <- nil
		}
	}()

	<-done

	log.Println("stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Printf("could not stop server cleanly: %v", err)
	}
}

// indWriter remplace chaque "\n" par ESC D (IND : descendre d'une ligne, sans
// retour chariot).
//
// Sur Windows, Bubble Tea suppose que "\n" ne ramène pas le curseur en colonne
// 0 (tea.go, mapNl = false). Or le terminal du client SSH ajoute un CR au LF :
// le curseur part en colonne 0, le "recule de N colonnes" qui suit est écrasé
// et la ligne est dessinée au mauvais endroit (résidus visuels).
type indWriter struct{ w io.Writer }

func (i indWriter) Write(p []byte) (int, error) {
	if _, err := i.w.Write(bytes.ReplaceAll(p, []byte("\n"), []byte("\x1bD"))); err != nil {
		return 0, err
	}
	return len(p), nil
}

// programHandler remplace bubbletea.Middleware : Wish ajoute ses options APRÈS
// celles du handler, donc un tea.WithOutput posé dans teaHandler serait écrasé.
func programHandler(s ssh.Session) *tea.Program {
	m, opts := teaHandler(s)
	opts = append(bubbletea.MakeOptions(s), opts...)
	if runtime.GOOS == "windows" {
		opts = append(opts, tea.WithOutput(indWriter{s}))
	}
	return tea.NewProgram(m, opts...)
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := model{
		width:  pty.Window.Width,
		height: pty.Window.Height,
	}

	// wish's bubbletea middleware only forces a color profile from $TERM on
	// unix builds (charm.land/wish/v2/bubbletea/tea_unix.go). On Windows the
	// server binary uses tea_other.go instead, which doesn't set one, so
	// styles render without color. Force it here from the client's env.
	envs := append(s.Environ(), "TERM="+pty.Term)

	return m, []tea.ProgramOption{
		tea.WithColorProfile(colorprofile.Env(envs)),
	}
}
