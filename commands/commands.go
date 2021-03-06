package commands

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/onodera-punpun/sponewm/focus"
	"github.com/onodera-punpun/sponewm/logger"
	"github.com/onodera-punpun/sponewm/wm"
	"github.com/onodera-punpun/sponewm/workspace"
	"github.com/onodera-punpun/sponewm/xclient"
)

// Env declares all available commands. Any command not in
// this list cannot be executed.
var Env = gribble.New([]gribble.Command{
	&Close{},
	&Focus{},
	&FocusRaise{},
	&FrameDecor{},
	&FrameNada{},
	&ToggleFloating{},
	&ToggleMaximize{},
	&ToggleStackAbove{},
	&ToggleStackBelow{},
	&ToggleSticky{},
	&Maximize{},
	&MouseMove{},
	&MouseResize{},
	&Move{},
	&MoveRelative{},
	&MovePointer{},
	&MovePointerRelative{},
	&Raise{},
	&Resize{},
	&Restart{},
	&Quit{},
	&Unmaximize{},
	&Workspace{},
	&WorkspaceSendClient{},
	&WorkspaceWithClient{},

	&Tile{},
	&Untile{},
	&TileToggle{},
	&MakeMaster{},

	&GetActive{},
	&GetAllClients{},
	&GetClientX{},
	&GetClientY{},
	&GetClientHeight{},
	&GetClientWidth{},
	&GetClientList{},
	&GetClientName{},
	&GetClientType{},
	&GetClientWorkspace{},
	&GetHead{},
	&GetNumHeads{},
	&GetNumHeadsConnected{},
	&GetHeadHeight{},
	&GetHeadWidth{},
	&GetHeadWorkspace{},
	&GetLayout{},
	&GetWorkspace{},
	&GetWorkspaceId{},
	&GetWorkspaceList{},
	&GetWorkspaceNext{},
	&GetWorkspacePrefix{},
	&GetWorkspacePrev{},
	&GetClientStatesList{},
	&HideClientFromPanels{},
	&ShowClientInPanels{},

	&TagGet{},
	&TagSet{},

	&True{},
	&False{},
	&MatchClientMapped{},
	&MatchClientClass{},
	&MatchClientInstance{},
	&MatchClientIsTransient{},
	&MatchClientName{},
	&MatchClientType{},
	&Not{},
	&And{},
	&Or{},
})

var (
	// SafeExec is a channel through which a Gribble command execution is
	// sent and executed synchronously with respect to the X main event loop.
	SafeExec = make(chan func() gribble.Value, 1)

	// SafeReturn is the means through which a return value from a Gribble
	// command is synchronously returned with respext to the X main event loop.
	// See SafeExec.
	SafeReturn = make(chan gribble.Value, 0)

	// Regex for enforcing tag name constraints.
	validTagName = regexp.MustCompile("^[-a-zA-Z0-9_]+$")
)

func init() {
	// This should be false in general for logging purposes.
	// When a command is executed via IPC, we temporarily turn it on so we
	// can give the user better error messages.
	Env.Verbose = false
}

var syncLock = new(sync.Mutex)

// syncRun should wrap the execution of most Gribble commands to ensure
// synchronous execution with respect to the main X event loop.
func syncRun(f func() gribble.Value) gribble.Value {
	syncLock.Lock()
	defer syncLock.Unlock()

	SafeExec <- f
	return <-SafeReturn
}

type Close struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Closes the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Close) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Close()
		})
		return nil
	})
}

type Focus struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Focuses the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Focus) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			if c == nil {
				focus.Root()

				// Use the mouse coordinates to find which workspace it was
				// clicked in. If a workspace can be found (i.e., no clicks in
				// dead areas), then activate it.
				xc, rw := wm.X.Conn(), wm.X.RootWin()
				qp, err := xproto.QueryPointer(xc, rw).Reply()
				if err != nil {
					logger.Warning.Printf("Could not query pointer: %s", err)
					return
				}

				geom := xrect.New(int(qp.RootX), int(qp.RootY), 1, 1)
				if wrk := wm.Heads.FindMostOverlap(geom); wrk != nil {
					wm.SetWorkspace(wrk, false)
				}
			} else {
				c.Focus()
				xevent.ReplayPointer(wm.X)
			}
		})
	})
}

type FocusRaise struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Focuses and raises the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FocusRaise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			c.Focus()
			c.Raise()
			xevent.ReplayPointer(wm.X)
		})
	})
}

type FrameDecor struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Set the decorations of the window specified by Client to the "Decor" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameDecor) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameDecor()
		})
		return nil
	})
}

type FrameNada struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Set the decorations of the window specified by Client to the "Nada" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameNada) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameNada()
		})
		return nil
	})
}

type ToggleFloating struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Toggles whether the window specified by Client should be forced into the
floating layout. A window forced into the floating layout CANNOT be tiled.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleFloating) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FloatingToggle()
		})
		return nil
	})
}

type ToggleMaximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Maximizes or restores the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleMaximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.MaximizeToggle()
		})
		return nil
	})
}

type ToggleStackAbove struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Toggles the layer of the window specified by Client from normal to above. When
a window is in the "above" layer, it will always be above other (normal)
clients.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleStackAbove) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StackAboveToggle()
		})
		return nil
	})
}

type ToggleStackBelow struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Toggles the layer of the window specified by Client from normal to below. When
a window is in the "below" layer, it will always be below other (normal)
clients.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleStackBelow) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StackBelowToggle()
		})
		return nil
	})
}

type ToggleSticky struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Toggles the sticky status of the window specified by Client. When a window is
sticky, it will always be visible unless iconified. (i.e., it does not belong
to any particular workspace.)

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleSticky) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StickyToggle()
		})
		return nil
	})
}

type Maximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Maximizes the window specified by Client. If the window is already maximized,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Maximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Maximize()
		})
		return nil
	})
}

type MouseMove struct {
	Help string `
Initiates a drag that allows a window to be moved with the mouse.

This is a special command that can only be assigned in SponeWM's mouse
configuration file. Invoking this command in any other way has no effect.
`
}

func (cmd MouseMove) Run() gribble.Value {
	logger.Warning.Printf("The MouseMove command can only be invoked from " +
		"the SponeWM mouse configuration file.")
	return nil
}

type MouseResize struct {
	Direction string `param:"1"`
	Help      string `
Initiates a drag that allows a window to be resized with the mouse.

Direction specifies how the window should be resized, and what the pointer
should look like. For example, if Direction is set to "BottomRight", then only
the width and height of the window can change---but not the x or y position.

Valid values for Direction are: Infer, Top, Bottom, Left, Right, TopLeft,
TopRight, BottomLeft and BottomRight. When "Infer" is used, the direction
is determined based on where the pointer is on the window when the drag is
initiated.

This is a special command that can only be assigned in SponeWM's mouse
configuration file. Invoking this command in any other way has no effect.
`
}

func (cmd MouseResize) Run() gribble.Value {
	logger.Warning.Printf("The MouseResize command can only be invoked from " +
		"the SpomeWM mouse configuration file.")
	return nil
}

type Raise struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Raises the window specified by Client to the top of its layer.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Raise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			c.Raise()
			xevent.ReplayPointer(wm.X)
		})
	})
}

type Move struct {
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
	Help   string      `
Moves the window specified by Client to the x and y position specified by
X and Y. Note that the origin is located in the top left corner.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Move) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		x, xok := parsePos(wm.Workspace().Geom(), cmd.X, false)
		y, yok := parsePos(wm.Workspace().Geom(), cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutMove(x, y)
		})
		return nil
	})
}

type MoveRelative struct {
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
	Help   string      `
Moves the window specified by Client to the x and y position specified by
X and Y, relative to its workspace. Note that the origin is located in the top
left corner of the client's workspace.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MoveRelative) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		geom := wm.Workspace().Geom()
		x, xok := parsePos(geom, cmd.X, false)
		y, yok := parsePos(geom, cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutMove(geom.X()+x, geom.Y()+y)
		})
		return nil
	})
}

type MovePointer struct {
	X    int    `param:"1"`
	Y    int    `param:"2"`
	Help string `
Moves the pointer to the x and y position specified by X and Y. Note the the
origin is located in the top left corner.
`
}

func (cmd MovePointer) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(cmd.X), int16(cmd.Y))
		return nil
	})
}

type MovePointerRelative struct {
	X    gribble.Any `param:"1" types:"int,float"`
	Y    gribble.Any `param:"2" types:"int,float"`
	Help string      `
Moves the pointer to the x and y position specified by X and Y relative to the
current workspace. Note the the origin is located in the top left corner of
the current workspace.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
workspace's geometry.
`
}

func (cmd MovePointerRelative) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		geom := wm.Workspace().Geom()
		x, xok := parsePos(geom, cmd.X, false)
		y, yok := parsePos(geom, cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(geom.X()+x), int16(geom.Y()+y))
		return nil
	})
}

type Restart struct {
	Help string `
Restarts SponeWM in place using exec. This should be used to reload SponeWM
after you've made changes to its configuration.
`
}

func (cmd Restart) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Restart = true // who says globals are bad?
		xevent.Quit(wm.X)
		return nil
	})
}

type Quit struct {
	Help string `
Stops SponeWM.
`
}

func (cmd Quit) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		logger.Message.Println("The User has told us to quit.")
		xevent.Quit(wm.X)
		return nil
	})
}

type Resize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Width  gribble.Any `param:"2" types:"int,float"`
	Height gribble.Any `param:"3" types:"int,float"`
	Help   string      `
Resizes the window specified by Client to some width and height specified by
Width and Height.

Width and Height may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Resize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		w, wok := parseDim(wm.Workspace().Geom(), cmd.Width, false)
		h, hok := parseDim(wm.Workspace().Geom(), cmd.Height, true)
		if !wok || !hok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutResize(w, h)
		})
		return nil
	})
}

type Unmaximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Unmaximizes the window specified by Client. If the window is not maximized,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Unmaximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Unmaximize()
		})
		return nil
	})
}

type Workspace struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Sets the current workspace to the one specified by Workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd Workspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wm.SetWorkspace(wrk, false)
			wm.FocusFallback()
		})
		return nil
	})
}

type WorkspaceSendClient struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Client    gribble.Any `param:"2" types:"int,string"`
	Help      string      `
Sends the window specified by Client to the workspace specified by Workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd WorkspaceSendClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				wrk.Add(c)
			})
		})
		return nil
	})
}

type WorkspaceWithClient struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Client    gribble.Any `param:"2" types:"int,string"`
	Help      string      `
Sets the current workspace to the workspace specified by Workspace, and moves
the window specified by Client to that workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd WorkspaceWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				c.Raise()
				wrk.Add(c)
				wm.SetWorkspace(wrk, false)
				wm.FocusFallback()
			})
		})
		return nil
	})
}

type TagGet struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Name   string      `param:"2"`
	Help   string      `
Retrieves the tag with name Name for the client specified by Client.

Client may be the window id or a substring that matches a window name.
Or, it may be zero and the property will be retrieved from the root
window.

Tag names may only contain the following characters: [-a-zA-Z0-9_].
`
}

func (cmd TagGet) Run() gribble.Value {
	if !validTagName.MatchString(cmd.Name) {
		return cmdError("Tag names must match %s.", validTagName.String())
	}

	var cid xproto.Window
	tagName := fmt.Sprintf("_SPONE_TAG_%s", cmd.Name)
	if n, ok := cmd.Client.(int); ok && n == 0 {
		cid = wm.Root.Id
	} else {
		withClient(cmd.Client, func(c *xclient.Client) {
			cid = c.Id()
		})
	}
	tval, err := xprop.PropValStr(xprop.GetProperty(wm.X, cid, tagName))
	if err != nil {
		// Log the error, but give the caller an empty string.
		logger.Warning.Println(err)
		return ""
	}
	return tval
}

type TagSet struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Name   string      `param:"2"`
	Value  string      `param:"3"`
	Help   string      `
Sets the tag with name Name to value Value for the client specified by Client.

Client may be the window id or a substring that matches a window name.
Or, it may be zero and the property will be set on the root window.

Tag names may only contain the following characters: [-a-zA-Z0-9_].
`
}

func (cmd TagSet) Run() gribble.Value {
	if !validTagName.MatchString(cmd.Name) {
		return cmdError("Tag names must match %s.", validTagName.String())
	}

	var cid xproto.Window
	tagName := fmt.Sprintf("_SPONE_TAG_%s", cmd.Name)
	if n, ok := cmd.Client.(int); ok && n == 0 {
		cid = wm.Root.Id
	} else {
		withClient(cmd.Client, func(c *xclient.Client) {
			cid = c.Id()
		})
	}
	err := xprop.ChangeProp(wm.X, cid, 8, tagName, "UTF8_STRING",
		[]byte(cmd.Value))
	if err != nil {
		return cmdError(err.Error())
	}
	return ""
}

type HideClientFromPanels struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Sets the appropriate flags so that the window specified by Client is
hidden from panels and pagers.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd HideClientFromPanels) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.SkipTaskbarSet(true)
			c.SkipPagerSet(true)
		})
		return nil
	})
}

type ShowClientInPanels struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help   string      `
Sets the appropriate flags so that the window specified by Client is
shown on panels and pagers.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ShowClientInPanels) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.SkipTaskbarSet(false)
			c.SkipPagerSet(false)
		})
		return nil
	})
}
