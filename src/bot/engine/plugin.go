package engine

import "log"
import "os"
import "encoding/json"
import "bytes"
import "fmt"
import "net/http"

type Plugin interface {
	Handle(b *Bot, req *BotRequest) (*BotResponse, bool)
	Name() string
	Init()
	Done()
	ConfigPath() string
	SetConfigPath(s string)
}

var plugins = make([]Plugin, 0)

func RegisterPlugin( p Plugin ) {
	plugins = append(plugins, p)
	log.Printf("Registered plugin %s", p.Name())
}

type PluginBase struct {
	Bot *Bot
	configPath string
}

func (b *PluginBase) ConfigPath() string {
	return b.configPath
}

func (b *PluginBase) SetConfigPath(s string) {
	b.configPath = s
	os.MkdirAll(b.configPath, 0700)
}

func (b *PluginBase) PostToIncoming( hookUrl string, payload *BotResponse ) error {

	bb, err := json.Marshal( payload )
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		hookUrl,
		bytes.NewBuffer(bb),
	)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode / 100 != 2 {
		return fmt.Errorf("Non 2xx response code returned: %d", resp.StatusCode)
	}

	return nil
}