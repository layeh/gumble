package barnard

import (
	"fmt"

	"github.com/bontibon/gumble/gumble"
)

func (b *Barnard) OnConnect(e *gumble.ConnectEvent) {
	b.Ui.SetActive(uiViewInput)
	b.UiTree.Rebuild()
	b.Ui.Refresh()

	b.UpdateInputStatus(fmt.Sprintf("To: %s", e.Client.Self().Channel().Name()))
	b.AddOutputLine(fmt.Sprintf("Connected to %s", b.Client.Conn().RemoteAddr()))
	if e.WelcomeMessage != "" {
		b.AddOutputLine(fmt.Sprintf("Welcome message: %s", esc(e.WelcomeMessage)))
	}
}

func (b *Barnard) OnDisconnect(e *gumble.DisconnectEvent) {
	b.AddOutputLine("Disconnected")
	b.UiTree.Rebuild()
	b.Ui.Refresh()
}

func (b *Barnard) OnTextMessage(e *gumble.TextMessageEvent) {
	b.AddOutputMessage(e.Sender, e.Message)
}

func (b *Barnard) OnUserChange(e *gumble.UserChangeEvent) {
	if e.ChannelChanged && e.User == b.Client.Self() {
		b.UpdateInputStatus(fmt.Sprintf("To: %s", e.User.Channel().Name()))
	}
	b.UiTree.Rebuild()
	b.Ui.Refresh()
}

func (b *Barnard) OnChannelChange(e *gumble.ChannelChangeEvent) {
	b.UiTree.Rebuild()
	b.Ui.Refresh()
}

func (b *Barnard) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	var info string
	switch e.Type {
	case gumble.PermissionDeniedOther:
		info = e.String
	case gumble.PermissionDeniedPermission:
		info = "insufficient permissions"
	case gumble.PermissionDeniedSuperUser:
		info = "cannot modify SuperUser"
	case gumble.PermissionDeniedInvalidChannelName:
		info = "invalid channel name"
	case gumble.PermissionDeniedTextTooLong:
		info = "text too long"
	case gumble.PermissionDeniedTemporaryChannel:
		info = "temporary channel"
	case gumble.PermissionDeniedMissingCertificate:
		info = "missing certificate"
	case gumble.PermissionDeniedInvalidUserName:
		info = "invalid user name"
	case gumble.PermissionDeniedChannelFull:
		info = "channel full"
	case gumble.PermissionDeniedNestingLimit:
		info = "nesting limit"
	}
	b.AddOutputLine(fmt.Sprintf("Permission denied: %s", info))
}
