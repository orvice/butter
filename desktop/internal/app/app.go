//go:build darwin

// Package app is the MyGo window: sign in, choose a workspace, show the
// connection. It holds no transport of its own; everything goes through
// account.Manager.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"go.orx.me/apps/butter/desktop/internal/account"
	"go.orx.me/apps/butter/desktop/internal/authapi"
	"go.orx.me/apps/butter/desktop/internal/credentials"
	"go.orx.me/apps/butter/desktop/internal/endpoint"
	"go.orx.me/apps/butter/desktop/internal/prefs"
	"go.orx.me/apps/butter/desktop/internal/version"
)

// Phase is what the window shows.
type Phase int

const (
	// PhaseLoading restores a stored session.
	PhaseLoading Phase = iota
	// PhaseSignedOut asks for a server and credentials.
	PhaseSignedOut
	// PhaseChooseWorkspace asks which workspace to use.
	PhaseChooseWorkspace
	// PhaseSignedIn shows the connection and the account.
	PhaseSignedIn
)

// Options configures the window.
type Options struct {
	// Prefs is the state saved by the previous run.
	Prefs prefs.Data
	// PrefsPath is where the window saves it; empty disables saving.
	PrefsPath string
	// Store keeps the session token.
	Store credentials.Store
	// Dial builds an API per server; nil uses the real ConnectRPC client.
	Dial func(apiBase string) account.API
}

// App is the window's state. It is read on the main thread only; background
// results come back through Window.Update.
type App struct {
	m         *account.Manager
	prefsPath string

	win  *mygo.Window
	opts Options

	phase    Phase
	server   string
	username string
	password string
	state    account.State
	busy     bool
	errText  string
	notice   string
}

// New returns an app for opts.
func New(opts Options) *App {
	if opts.Store == nil {
		opts.Store = credentials.NewSystem("go.orx.me/apps/butter/desktop")
	}
	return &App{
		m:         account.New(opts.Store, opts.Dial),
		prefsPath: opts.PrefsPath,
		opts:      opts,
		phase:     PhaseSignedOut,
		server:    opts.Prefs.Server,
		username:  opts.Prefs.Username,
		state:     account.State{Server: opts.Prefs.Server},
	}
}

// Start creates the window and restores a stored session. It runs on the
// main thread, from mygo.App.WhenReady.
func (a *App) Start() {
	a.win = mygo.NewWindow(mygo.WindowOptions{
		Title:   "Butter Desktop",
		Width:   460,
		Height:  560,
		Content: ui.View(a.view),
	})
	a.restore()
}

// restore signs in again from the Keychain, if the previous run left a
// session there.
func (a *App) restore() {
	server := a.server
	if server == "" {
		return
	}
	a.busy = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := a.m.UseServer(server); err != nil {
			a.update(func() {
				a.busy = false
				a.errText = message(err)
			})
			return
		}
		st, err := a.m.Restore(ctx)
		a.update(func() {
			a.busy = false
			switch {
			case err == nil:
				a.apply(st)
			case errors.Is(err, account.ErrNoSession):
				a.phase = PhaseSignedOut
			case authapi.Is(err, authapi.KindSessionExpired):
				a.phase = PhaseSignedOut
				a.errText = "登录已过期，请重新登录"
			case errors.Is(err, account.ErrStale):
				a.phase = PhaseSignedOut
			default:
				a.phase = PhaseSignedOut
				a.errText = message(err)
			}
		})
	}()
}

// login runs in the background; the result is applied on the main thread.
func (a *App) login() {
	if a.busy {
		return
	}
	server, username, password := a.server, a.username, a.password
	a.busy = true
	a.errText = ""
	a.notice = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if _, err := a.m.UseServer(server); err != nil {
			a.update(func() {
				a.busy = false
				a.errText = message(err)
			})
			return
		}
		st, err := a.m.SignIn(ctx, username, password)
		a.update(func() {
			a.busy = false
			a.password = ""
			if err != nil {
				if errors.Is(err, account.ErrStale) {
					a.errText = "服务器已切换，请重新登录"
					return
				}
				a.errText = message(err)
				return
			}
			a.apply(st)
		})
	}()
}

// signOut forgets the session here and revokes it on the server when it can.
func (a *App) signOut() {
	if a.busy {
		return
	}
	a.busy = true
	a.errText = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := a.m.SignOut(ctx)
		a.update(func() {
			a.busy = false
			a.state = account.State{Server: a.server}
			a.phase = PhaseSignedOut
			a.notice = "已退出登录"
			if err != nil {
				a.notice = "已在本地退出；服务端会话未能撤销"
			}
		})
	}()
}

// selectWorkspace is local: the server's membership list is the limit.
func (a *App) selectWorkspace(id string) {
	st, err := a.m.SelectWorkspace(id)
	if err != nil {
		a.errText = message(err)
		return
	}
	a.errText = ""
	a.apply(st)
}

func (a *App) apply(st account.State) {
	a.state = st
	a.errText = ""
	switch {
	case !st.SignedIn():
		a.phase = PhaseSignedOut
	case st.WorkspaceID == "" && len(st.Workspaces) > 0:
		a.phase = PhaseChooseWorkspace
	case st.WorkspaceID == "":
		a.phase = PhaseSignedIn
		a.notice = "该账号没有可访问的工作区"
	default:
		a.phase = PhaseSignedIn
	}
	a.savePrefs()
}

func (a *App) savePrefs() {
	if a.prefsPath == "" {
		return
	}
	err := prefs.Save(a.prefsPath, prefs.Data{
		Server:      a.state.Server,
		Username:    a.state.User.Username,
		WorkspaceID: a.state.WorkspaceID,
	})
	if err != nil {
		a.notice = "偏好设置未能保存"
	}
}

func (a *App) update(fn func()) {
	if a.win == nil {
		fn()
		return
	}
	a.win.Update(fn)
}

func (a *App) view(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Padding(24).Gap(12).Children(func() {
		ui.Text(c, "Butter Desktop").FontSize(20).Bold()
		ui.Text(c, a.subtitle()).FontSize(12).TextColor(t.TextMuted)

		switch a.phase {
		case PhaseLoading:
			ui.Text(c, "正在恢复会话…")
		case PhaseSignedOut:
			a.viewSignedOut(c)
		case PhaseChooseWorkspace:
			a.viewWorkspaces(c)
		case PhaseSignedIn:
			a.viewSignedIn(c)
		}

		if a.errText != "" {
			ui.Text(c, a.errText).FontSize(12).TextColor(t.Danger)
		}
		if a.notice != "" {
			ui.Text(c, a.notice).FontSize(12).TextColor(t.TextMuted)
		}
		ui.Text(c, version.String()).FontSize(11).TextColor(t.TextMuted)
	})
}

func (a *App) subtitle() string {
	switch a.phase {
	case PhaseLoading:
		return "正在连接"
	case PhaseChooseWorkspace:
		return "选择工作区"
	case PhaseSignedIn:
		if a.state.User.Username != "" {
			return a.state.User.Username
		}
		return "已登录"
	default:
		return "登录到你的 Butter 服务器"
	}
}

func (a *App) viewSignedOut(c *ui.Context) {
	ui.Column(c).Gap(10).Children(func() {
		ui.TextInput(c, &a.server).
			Placeholder("https://butter.example.com").
			Label("服务器地址").
			Width(360).
			Disabled(a.busy)
		if ui.TextInput(c, &a.username).
			Placeholder("用户名").
			Label("用户名").
			Width(360).
			AutoFocus().
			Disabled(a.busy).
			Submitted() {
			a.login()
		}
		if ui.TextInput(c, &a.password).
			Placeholder("密码").
			Label("密码").
			Password().
			Width(360).
			Disabled(a.busy).
			Submitted() {
			a.login()
		}
		ui.Row(c).Gap(8).Children(func() {
			ui.PrimaryButton(c, "登录").Disabled(a.busy).OnClick(a.login)
		})
		ui.Text(c, "密码只用于这一次登录，不会保存在本机；会话 token 存在系统钥匙串里。").
			FontSize(11).
			TextColor(c.Theme().TextMuted)
	})
}

func (a *App) viewWorkspaces(c *ui.Context) {
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, "该账号可以访问多个工作区，请选择一个：").FontSize(12)
		for _, w := range a.state.Workspaces {
			ws := w
			label := ws.Name
			if ws.Slug != "" && ws.Slug != ws.Name {
				label = fmt.Sprintf("%s（%s）", ws.Name, ws.Slug)
			}
			ui.Button(c.Key(ws.ID), label).OnClick(func() { a.selectWorkspace(ws.ID) })
		}
		ui.Button(c, "退出登录").OnClick(a.signOut)
	})
}

func (a *App) viewSignedIn(c *ui.Context) {
	t := c.Theme()
	name := a.state.User.DisplayName
	if name == "" {
		name = a.state.User.Username
	}
	ui.Column(c).Gap(8).Children(func() {
		ui.Textf(c, "已登录：%s", name)
		ui.Textf(c, "服务器：%s", a.state.Server).FontSize(12).TextColor(t.TextMuted)
		if ws := a.workspaceLabel(); ws != "" {
			ui.Textf(c, "工作区：%s", ws).FontSize(12).TextColor(t.TextMuted)
		}
		if !a.state.ExpiresAt.IsZero() {
			ui.Textf(c, "会话到期：%s", a.state.ExpiresAt.Local().Format("2006-01-02 15:04")).
				FontSize(12).
				TextColor(t.TextMuted)
		}
		ui.Text(c, "聊天功能在下一个里程碑接入；当前版本只验证登录与会话恢复。").
			FontSize(11).
			TextColor(t.TextMuted)
		ui.Row(c).Gap(8).Children(func() {
			ui.Button(c, "退出登录").Disabled(a.busy).OnClick(a.signOut)
		})
	})
}

func (a *App) workspaceLabel() string {
	for _, w := range a.state.Workspaces {
		if w.ID == a.state.WorkspaceID {
			if w.Name != "" {
				return w.Name
			}
			return w.Slug
		}
	}
	return a.state.WorkspaceID
}

// message turns a failure into text for the window, without server
// internals or anything secret.
func message(err error) string {
	var apiErr *authapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Message()
	}
	switch {
	case errors.Is(err, endpoint.ErrInvalid):
		return "服务器地址无效，例如 https://butter.example.com"
	case errors.Is(err, account.ErrNoServer):
		return "请先填写服务器地址"
	case errors.Is(err, account.ErrNoSession):
		return "没有已保存的登录状态，请重新登录"
	case errors.Is(err, account.ErrUnknownWorkspace):
		return "该工作区不可用，请重新选择"
	case errors.Is(err, credentials.ErrUnsupported):
		return "系统没有可用的凭据存储，无法保存登录状态"
	case errors.Is(err, context.DeadlineExceeded):
		return "请求超时，请检查网络"
	case errors.Is(err, context.Canceled):
		return "已取消"
	}
	return "操作失败，请重试"
}
