package templater

import (
	"BangumiBot/data"
	"bytes"
	"text/template"
	"time"
)

var temps = template.Must(template.New("season_templates").
	Funcs(template.FuncMap{"now": time.Now}).
	ParseGlob("template/*"))

func Season(s data.Season) string {
	var buffer bytes.Buffer
	err := temps.ExecuteTemplate(&buffer, "season.txt", s)
	if err != nil {
		panic(err)
	} else {
		return buffer.String()
	}
}

func QueryReply(sli []data.Season) string {
	var buffer bytes.Buffer
	err := temps.ExecuteTemplate(&buffer, "query_reply.txt", sli)
	if err != nil {
		panic(err)
	} else {
		return buffer.String()
	}
}

func PubNotice(s data.Season) string {
	var buffer bytes.Buffer
	err := temps.ExecuteTemplate(&buffer, "pub_notice.txt", s)
	if err != nil {
		panic(err)
	} else {
		return buffer.String()
	}
}
