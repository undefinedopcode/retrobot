package engine

import "os"
import "bot/config"
import "net/http"
import "fmt"
import "log"
import "io/ioutil"
import "strconv"
import "github.com/gorilla/mux"
import "encoding/json"
import "errors"
import "strings"
import "regexp"

func New(cfg *config.Config) (*Bot, error) {
	if cfg.Username == "" {
		return nil, errors.New("UserName required")
	}
	if cfg.Token == "" {
		return nil, errors.New("Mattermost webhook token required")
	}
	if cfg.Port == 0 {
		return nil, errors.New("Port required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "localhost"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = fmt.Sprintf("./data/%s", cfg.Username)
		os.MkdirAll(cfg.DataDir, 0700)
	}
	bot := &Bot{
		Config: cfg,
	}	
	bot.Init()
	return bot, bot.Start()
}

func (b *Bot) Start() error {
	r := mux.NewRouter()
	r.HandleFunc( "/message", b.Message )
	r.HandleFunc( "/slash/{command}", b.Slash)
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))),
	)
	
	var err error
	go func() {
		err = http.ListenAndServe( 
			fmt.Sprintf("%s:%d", b.Config.BaseURL, b.Config.Port),
			r,
		)
	}()

	if err == nil {
		log.Printf("Retrobot is available on %s:%d", b.Config.BaseURL, b.Config.Port)
	}
			
	return err
}

func (b *Bot) Slash( w http.ResponseWriter, r *http.Request ) {
	if r.Method == "POST" {
		r.ParseForm()
		bb, err := ioutil.ReadAll(r.Body)
		if err == nil {
			log.Printf("Request with content-type: %s", r.Header.Get("Content-Type"))
			log.Printf("Request body: %s", string(bb))
			var req BotRequest
			switch r.Header.Get("Content-Type") {
			case "application/json":
				err = json.Unmarshal(bb, &req)
				if err != nil {
					return
				}
			case "application/x-www-form-urlencoded":
				req.Token = r.Form.Get("token")
				req.UserName = r.Form.Get("user_name")
				req.UserID = r.Form.Get("user_id")
				req.TriggerWord = r.Form.Get("trigger_word")
				req.ChannelID = r.Form.Get("channel_id")
				req.ChannelName = r.Form.Get("channel_name")
				req.TeamDomain = r.Form.Get("team_domain")
				req.TeamID = r.Form.Get("team_id")
				req.PostID = r.Form.Get("post_id")
				req.Text = r.Form.Get("text")
				req.Timestamp, _ = strconv.ParseInt(r.Form.Get("timestamp"), 10, 64)				
			default:
				log.Printf("Unknown request type. Skipping")
			}
			if b.Config.IsTokenValid(true, req.Token) {
				// request ok
				log.Println("Token was valid")
			} else {
				log.Println("Ignoring invalid token")
				return
			}
			log.Printf("Parsed request is : %+v", req)

			// in the base of a slash, reappend things
			req.Text = mux.Vars(r)["command"] + " " + req.Text

			resp := b.HandleRequest( &req )
			if resp != nil {
				if resp.ResponseType == "" {
					resp.ResponseType = "in_channel"
				}
				bb, err = json.Marshal(resp)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(bb)
					return
				}			
			}
		}		
	} else {
		log.Printf("Invalid request method: %s", r.Method)
	}
}

func (b *Bot) Message( w http.ResponseWriter, r *http.Request ) {

	if r.Method == "POST" {
		r.ParseForm()
		bb, err := ioutil.ReadAll(r.Body)
		if err == nil {
			log.Printf("Request with content-type: %s", r.Header.Get("Content-Type"))
			log.Printf("Request body: %s", string(bb))
			var req BotRequest
			switch r.Header.Get("Content-Type") {
			case "application/json":
				err = json.Unmarshal(bb, &req)
				if err != nil {
					return
				}
			case "application/x-www-form-urlencoded":
				req.Token = r.Form.Get("token")
				req.UserName = r.Form.Get("user_name")
				req.UserID = r.Form.Get("user_id")
				req.TriggerWord = r.Form.Get("trigger_word")
				req.ChannelID = r.Form.Get("channel_id")
				req.ChannelName = r.Form.Get("channel_name")
				req.TeamDomain = r.Form.Get("team_domain")
				req.TeamID = r.Form.Get("team_id")
				req.PostID = r.Form.Get("post_id")
				req.Text = r.Form.Get("text")
				req.Timestamp, _ = strconv.ParseInt(r.Form.Get("timestamp"), 10, 64)				
			default:
				log.Printf("Unknown request type. Skipping")
			}
			if b.Config.IsTokenValid(false, req.Token) {
				// request ok
				log.Println("Token was valid")
			} else {
				log.Println("Ignoring invalid token")
				return
			}
			log.Printf("Parsed request is : %+v", req)

			resp := b.HandleRequest( &req )
			if resp != nil {
				bb, err = json.Marshal(resp)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(bb)
					return
				}			
			}
		}
	} else {
		log.Printf("Invalid request method %s", r.Method)
	}
	
}

var reField = regexp.MustCompile("([$][{]([^${}]+)[}])")

func (b *Bot) Expand( value string ) string {
	return b.expand(value)
}

func (b *Bot) expand( value string ) string {
	for reField.MatchString(value) {
		m := reField.FindAllStringSubmatch(value, -1)
		fullPattern := m[0][1]
		name := m[0][2]
		log.Printf("Got field match %s (%s)", fullPattern, name)
		var replacement string
		switch name {
		case "baseurl":
			replacement = b.Config.BaseURL
		case "port":
			replacement = fmt.Sprintf("%d", b.Config.Port)
		}
		log.Printf("Replace %s with %s", fullPattern, replacement)
		value = strings.Replace(value, fullPattern, replacement, -1)
	}
	return value
}

func (b *Bot) HandleRequest( req *BotRequest ) *BotResponse {
	for _, p := range b.Plugins {
		resp, ok := p.Handle(b, req)
		if ok && resp != nil {
			if resp.UserName == "" {
				resp.UserName = b.Config.Username
			}
			if resp.IconURL == "" {
				resp.IconURL = b.expand(b.Config.IconURL)
			}
			return resp
		}
	}	
	return nil
}

func (b *Bot) Init() {
	b.Plugins = GetPlugins(b)
	b.InitPlugins()
}

func (b *Bot) Done() {
	b.DonePlugins()
}

func (b *Bot) InitPlugins() {
	for _, p := range b.Plugins {
		p.SetConfigPath( b.Config.DataDir + "/" + p.Name() )
		p.Init()
	}
}

func (b *Bot) DonePlugins() {
	for _, p := range b.Plugins {
		p.Done()
	}	
}