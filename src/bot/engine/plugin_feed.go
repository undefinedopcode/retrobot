package engine

import "time"
import "log"
import "io/ioutil"
import "gopkg.in/yaml.v2"
import "github.com/mmcdole/gofeed"

type Feed struct {
	Name string
	URL string
	CheckMinutes int
	Hooks []string
	lastMaxID int
	lastUpdated time.Time
	lastTime time.Time
}

type PluginFeedConfig struct {
	FeedList []*Feed
	AuthUserID []string
	MustBeAuthorized bool
}

type PluginFeed struct {
	PluginBase
	Config *PluginFeedConfig
	fp *gofeed.Parser
}

func init() {
}

func (p *PluginFeed) Init() {
	log.Printf("Init for plugin %s", p.Name()) 	 
	b, err := ioutil.ReadFile( p.ConfigPath()+"/config.yml" )
	if err == nil {
		log.Println("Found feed config..")
		cfg := &PluginFeedConfig{}
		err = yaml.Unmarshal(b, cfg)
		if err == nil {
			log.Printf("Parsed feed config and got %d feeds", len(cfg.FeedList))
			p.Config = cfg	
			p.FetchAndUpdate(false)
//			p.FetchAndUpdate(false)
		} else {
			log.Printf("Parsing feed config failed: %v", err)
		}
	} else {
		log.Printf("Reading feed config failed: %v", err)
	}
	go func() {
		for {
			time.Sleep(time.Minute)
			p.FetchAndUpdate(true)
		}
	}()
}

func (p *PluginFeed) FetchAndUpdate(broadcast bool) {
	if p.fp == nil {
		p.fp = gofeed.NewParser()
	}
	for _, feed := range p.Config.FeedList {
		if time.Since(feed.lastTime) > time.Duration(feed.CheckMinutes) * time.Minute {
			feed.lastTime = time.Now()
			f, _ := p.fp.ParseURL(feed.URL)
			log.Printf("Updating feed %s", f.Title)
			updates := make([]*gofeed.Item, 0, 20)
			for i:=len(f.Items)-1; i>=0; i-- {
				item := f.Items[i]
				if item.PublishedParsed != nil && item.PublishedParsed.After( feed.lastUpdated ) {
					log.Printf("Update found: %s", item.Title)
					feed.lastUpdated = *item.PublishedParsed
					updates = append(updates, item)
				}
			}
			log.Printf("Got %d updates", len(updates))
			if len(updates) > 0 && broadcast {
				//updates = updates[len(updates)-1:]
				for _, item := range updates {
					for _, hook := range feed.Hooks {
						log.Printf("POST %s", hook)
						err := p.PostToIncoming( 
							hook,
							&BotResponse{
								UserName: p.Bot.Config.Username,
								IconURL: p.Bot.Expand(p.Bot.Config.IconURL),
								Text: feed.Name + ": " + item.Link,		
							},
						)
						if err != nil {
							log.Printf("POST failed with error: %v", err)
						}
					}
				}
			}
		}
	}
}

func (p *PluginFeed) Done() {
	//
}

func (p *PluginFeed) Name() string {
	return "Feed"
}

func (p *PluginFeed) Handle( b *Bot, req *BotRequest ) (*BotResponse, bool) {

	command, _ := req.CommandAndArgs(0)

	if command != "feed" {
		return nil, false
	}

	r := &BotResponse{
		//Text: fmt.Sprintf("You roll %d", rand.Intn(6)+1),
	}
	r.AddAttachment(
		&BotResponseAttachment{
			Color: "#ff0000",	
			Text:  "feed thing",
		},
	)
	return r, true
}

func NewPluginFeed(b *Bot) *PluginFeed {
	return &PluginFeed{
		Config: &PluginFeedConfig{},
		PluginBase: PluginBase{
			Bot: b,
		},
	}
}

