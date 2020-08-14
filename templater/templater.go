package templater

import (
	"BangumiBot/data"
	"bytes"
	"text/template"
	"time"
)

type MessageTemplate template.Template

func LoadGlob(glob string) *MessageTemplate {
	return (*MessageTemplate)(template.Must(template.New("season_templates").
		Funcs(template.FuncMap{"now": time.Now}).
		ParseGlob(glob)))
}

func (t *MessageTemplate) executeTemplate(filename string, data interface{}) string {
	var buffer bytes.Buffer
	err := (*template.Template)(t).ExecuteTemplate(&buffer, filename, data)
	if err != nil {
		panic(err)
	} else {
		return buffer.String()
	}
}

func (t *MessageTemplate) Season(s data.Season) string {
	return t.executeTemplate("season.txt", s)
}

func (t *MessageTemplate) QueryReply(sli []data.Season) string {
	return t.executeTemplate("query_reply.txt", sli)
}

func (t *MessageTemplate) PubNotice(s data.Season) string {
	return t.executeTemplate("pub_notice.txt", s)
}
