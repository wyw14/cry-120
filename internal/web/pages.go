package web

import (
	"errors"
	"html/template"
	"io"
)

type Page struct {
	Title    string
	Subtitle string
	Endpoint string
	Actions  template.HTML
}

type Pages struct {
	template *template.Template
	items    map[string]Page
}

func NewPages() *Pages {
	markup := `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}} | LaunchGuard</title><style>` + stylesheet + `</style></head><body data-endpoint="{{.Endpoint}}"><header><strong>LAUNCHGUARD</strong><nav><a href="/launch">倒计时</a><a href="/propellant">推进剂</a><a href="/umbilical">脐带塔</a><a href="/safety">安全联锁</a></nav></header><main><h1>{{.Title}}</h1><p class="subtitle">{{.Subtitle}}</p><div class="toolbar">{{.Actions}}</div><div id="statusline" class="statusline warn">Connecting</div><div class="grid"><section class="panel"><h2>Live control state</h2><pre id="output">Loading</pre></section><section class="panel"><h2>Operational boundary</h2><p>Commands are accepted only while their generation, revision, feedback and resource fencing remain current.</p></section></div></main><script>` + script + `</script></body></html>`
	tmpl := template.Must(template.New("page").Parse(markup))
	return &Pages{template: tmpl, items: map[string]Page{
		"/launch":     {Title: "发射倒计时", Subtitle: "查看 generation、稳定等待和当前禁令。", Endpoint: "/api/countdown", Actions: `<button class="primary" data-action="resume">恢复倒计时</button>`},
		"/propellant": {Title: "推进剂加注", Subtitle: "协调燃料、氧化剂与共享转输歧管。", Endpoint: "/api/propellant", Actions: `<button data-action="fill" data-kind="fuel" data-arm="fuel-arm-1">燃料补加</button><button data-action="fill" data-kind="oxidizer" data-arm="lox-arm-1">氧化剂补加</button>`},
		"/umbilical":  {Title: "脐带塔撤收", Subtitle: "关联动作 token、设备反馈与控制器 quorum。", Endpoint: "/api/umbilical", Actions: `<button class="primary" data-action="arm">准备撤收</button>`},
		"/safety":     {Title: "安全联锁", Subtitle: "汇总来源级禁令、许可与安全结论。", Endpoint: "/api/interlocks", Actions: `<button data-action="hold">施加人工禁令</button>`},
	}}
}

func (p *Pages) Render(path string, writer io.Writer) error {
	page, ok := p.items[path]
	if !ok {
		return errors.New("page not found")
	}
	return p.template.Execute(writer, page)
}

func (p *Pages) Paths() []string {
	return []string{"/launch", "/propellant", "/umbilical", "/safety"}
}
