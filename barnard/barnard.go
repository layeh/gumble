package barnard

import (
	"github.com/bontibon/gumble/barnard/uiterm"
	"github.com/bontibon/gumble/gumble"
	"github.com/bontibon/gumble/gumble_openal"
)

type Barnard struct {
	Config gumble.Config
	Client *gumble.Client

	Stream *gumble_openal.Stream

	Ui            *uiterm.Ui
	UiOutput      uiterm.Textview
	UiInput       uiterm.Textbox
	UiStatus      uiterm.Label
	UiTree        uiterm.Tree
	UiInputStatus uiterm.Label
}
