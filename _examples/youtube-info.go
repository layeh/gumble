package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)

const responseTemplate = `
<table>
    <tr>
        <td valign="middle">
            <img src='https://www.youtube.com/yt/brand/media/image/YouTube-icon-full_color.png' height="25" />
        </td>
        <td align="center" valign="middle">
            <a href="http://youtu.be/{{ .Data.Id }}">{{ .Data.Title }} ({{ .Data.Duration }})</a>
        </td>
    </tr>
    <tr>
        <td></td>
        <td align="center">
            <a href="http://youtu.be/{{ .Data.Id }}"><img src="{{ .Data.Thumbnail.HqDefault }}" width="250" /></a>
        </td>
    </tr>
</table>`

const linkPattern = `https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/v/|youtube\.com/v/)([[:alnum:]_\-]+)`

type videoInfo struct {
	Data struct {
		Id        string
		Title     string
		Duration  time.Duration
		Thumbnail struct {
			HqDefault string
		}
	}
}

var pattern *regexp.Regexp
var outputTemplate *template.Template

func init() {
	var err error
	pattern = regexp.MustCompile(linkPattern)
	outputTemplate, err = template.New("root").Parse(responseTemplate)
	if err != nil {
		panic(err)
	}
}

func fetchYoutubeInfo(client *gumble.Client, id string) {
	var info videoInfo

	// Fetch + parse video info
	url := fmt.Sprintf("http://gdata.youtube.com/feeds/api/videos/%s?v=2&alt=jsonc", id)
	if resp, err := http.Get(url); err != nil {
		return
	} else {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&info); err != nil {
			return
		}
		info.Data.Duration *= time.Second
		resp.Body.Close()
	}

	// Create response string
	var buffer bytes.Buffer
	if err := outputTemplate.Execute(&buffer, info); err != nil {
		return
	}
	message := gumble.TextMessage{
		Channels: []*gumble.Channel{
			client.Self.Channel,
		},
		Message: buffer.String(),
	}
	client.Send(&message)
}

func main() {
	gumbleutil.Main(nil, gumbleutil.Listener{
		// Connect event
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Printf("youtube-info loaded!\n")
		},

		// Text message event
		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			matches := pattern.FindStringSubmatch(e.Message)
			if len(matches) != 2 {
				return
			}
			go fetchYoutubeInfo(e.Client, matches[1])
		},
	})
}
