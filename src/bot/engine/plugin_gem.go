package engine

import "os"
import "io/ioutil"
import "log"
import "time"
import "fmt"
import "math/rand"
import "strconv"
import "strings"
import "sync"
import "gopkg.in/yaml.v2"

type Gem struct {
	ID int
	Creator string
	Date time.Time
	Text string
}

type GemDB struct {
	m sync.Mutex
	fn string
	Count int
	Gems map[string][]*Gem
}

func (db *GemDB) Add( channelid string, creator string, date time.Time, text string ) (int, error) {
	db.m.Lock()
	defer db.m.Unlock()
	nextId := db.Count
	db.Count++

	g := &Gem{
		ID: nextId,
		Creator: creator,
		Date: date,
		Text: text,
	} 

	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}
	list = append(list, g)
	db.Gems[channelid] = list

	return nextId, db.Save()	
}

func (db *GemDB) Get(channelid string, id int) *Gem {
	db.m.Lock()
	defer db.m.Unlock()
	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}	
	for _, g := range list {
		if g.ID == id {
			return g
		}
	}	
	return nil
}

func (db *GemDB) Random(channelid string) *Gem {
	db.m.Lock()
	defer db.m.Unlock()
	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}		
	if len(list) == 0 {
		return nil
	}
	return list[rand.Intn(len(list))]
}

func (db *GemDB) GetCount(channelid string) int {
	db.m.Lock()
	defer db.m.Unlock()
	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}	
	return len(list)
}

func (db *GemDB) Search(channelid string, term string, max int) []*Gem {
	db.m.Lock()
	defer db.m.Unlock()
	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}	
	out := make([]*Gem, 0)	
	for _, g := range list {
		if strings.Contains(strings.ToLower(g.Creator), strings.ToLower(term)) || 
		   strings.Contains(strings.ToLower(g.Text), strings.ToLower(term)) {
			out = append(out, g)   	
		}
		if max != -1 && len(out) >= max {
			break
		}
	}	
	return out
}

func (db *GemDB) Remove(channelid string, id int) error {
	db.m.Lock()
	defer db.m.Unlock()
	list, ok := db.Gems[channelid]
	if !ok {
		list = make([]*Gem, 0)
	}	
	out := make([]*Gem, 0)	
	for _, g := range list {
		if g.ID != id {
			out = append(out, g)
		}
	}
	if len(out) == len(list) {
		return fmt.Errorf("No such gem %d in channel", id)
	}	
	db.Gems[channelid] = out
	return db.Save()
}

func (db *GemDB) Save() error {
	f, err := os.Create(db.fn)
	if err != nil {
		return err
	}
	defer f.Close()
	y, err := yaml.Marshal(db)
	if err != nil {
		return err
	}
	f.Write(y)
	log.Printf("Saved %s", db.fn)	
	return nil
}

func (db *GemDB) Load() error {
	b, err := ioutil.ReadFile(db.fn)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, db)
	if err == nil {
		log.Printf("Loaded %s", db.fn)
	}
	return err	
}

func NewGemDB(fn string) (*GemDB, error) {
	d := &GemDB{
		fn: fn,
		Gems: make(map[string][]*Gem),
	}
	err := d.Load()
	return d, err
}

type PluginGem struct {
	PluginBase
	db *GemDB
}

func (p *PluginGem) Init() {
	//
	log.Printf("Init for plugin %s", p.Name()) 
	datafile := p.ConfigPath() + "/gems.yml"
	p.db, _ = NewGemDB(datafile)
}

func (p *PluginGem) Done() {
	//
}

func (p *PluginGem) Name() string {
	return "Gem"
}

func (p *PluginGem) Handle( b *Bot, req *BotRequest ) (*BotResponse, bool) {

	command, args := req.CommandAndArgs(2)

	log.Printf("Command: %s, Args: %+v", command, args)
	
	if command != "gem" {
		return nil, false
	}

	channelid := req.ChannelID
	text := ""
	title := "Gems"

	r := &BotResponse{
	}

	if len(args) == 0 {
		// do gem
		g := p.db.Random(req.ChannelID)
		if g == nil {
			text = "No gems for this channel.  Add one with /gem add ..."
		} else {
			title = fmt.Sprintf("#%d (posted by %s on %v)", g.ID, g.Creator, g.Date)
			text = g.Text
		}
	} else {
		switch args[0] {
		case "help":
			text = "```Help:\n" +
				    "/gem add <text>    Adds a gem.\n" +
				    "/gem remove <id>   Remove a gem if you own it.\n" +
				    "/gem <id>          Show specific gem.\n" +
				    "/gem               Show a random gem.```\n"
		case "add":
			t := ""
			if len(args) > 1 {
				t = args[1]
			}
			if t != "" {
				id, err := p.db.Add(req.ChannelID, req.UserName, time.Now(), t)
				if err == nil {
					title = fmt.Sprintf("#%d (posted by %s on %v)", id, req.UserName, time.Now())
					text = t
				} else {
					text = "Failed to add gem."
				}
			}
		case "remove":
			// maybe a number
			i, _ := strconv.ParseInt(args[1], 10, 64)
			g := p.db.Get(channelid, int(i))
			if g == nil {
				text = "No such gem.  Add one with /gem add ..."
			} else {
				if g.Creator != req.UserName {
					text = "Not owner of gem."
				} else {
					err := p.db.Remove(channelid, int(i))
					if err == nil {
						text = "Removed gem."
					} else {
						text = fmt.Sprintf("Failed to remove gem: %v", err)
					}
				}
			}
		default: 
			// maybe a number
			i, _ := strconv.ParseInt(args[0], 10, 64)
			g := p.db.Get(channelid, int(i))
			if g == nil {
				text = "No such gem.  Add one with /gem add ..."
			} else {
				title = fmt.Sprintf("#%d (posted by %s on %v)", g.ID, g.Creator, g.Date)
				text = g.Text
			}
		}
	}

	r.AddAttachment(
		&BotResponseAttachment{
			Color: "#0000ff",	
			Text: text,
			Title: title,
		},
	)
	return r, true
}

func NewPluginGem(b *Bot) *PluginGem {
	return &PluginGem{
		PluginBase: PluginBase{
			Bot: b,
		},
	}
}
