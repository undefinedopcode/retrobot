package engine

import "log"
import "time"
import "fmt"
import "math/rand"
import "regexp"
import "strconv"
import "strings"

type PluginDice struct {
	PluginBase
}

func init() {
}

var reNSided = regexp.MustCompile("^([0-9]+)$")
var rexDy = regexp.MustCompile("^([0-9]+)?d([0-9]+)$")

func (p *PluginDice) Init() {
	//
	log.Printf("Init for plugin %s", p.Name()) 
	rand.Seed(time.Now().UnixNano())
}

func (p *PluginDice) Done() {
	//
}

func (p *PluginDice) Name() string {
	return "Dice"
}

func (p *PluginDice) Handle( b *Bot, req *BotRequest ) (*BotResponse, bool) {

	command, args := req.CommandAndArgs(1)

	log.Printf("Command: %s, Args: %+v", command, args)
	
		if command != "roll" {
		return nil, false
	}

	qty := 1
	sides := 6

	if len(args) == 1 {
		if reNSided.MatchString(args[0]) {
			m := reNSided.FindAllStringSubmatch(args[0], -1)
			i, _ := strconv.ParseInt(m[0][1], 10, 64)
			sides = int(i)
		} else if rexDy.MatchString(args[0]) {
			m := rexDy.FindAllStringSubmatch(args[0], -1)
			i, _ := strconv.ParseInt(m[0][1], 10, 64)
			qty = int(i)
			i, _ = strconv.ParseInt(m[0][2], 10, 64)
			sides = int(i)
		}
	}

	if sides < 1 {
		sides = 1
	}
	if qty < 1 {
		qty = 1
	}
	if qty > 20 {
		qty = 20
	}
	
	results := make([]string, qty)
	for i, _ := range results {
		results[i] = fmt.Sprintf( "%d", rand.Intn(sides)+1 )
	}	

	r := &BotResponse{
		//Text: fmt.Sprintf("You roll %d", rand.Intn(6)+1),
	}
	r.AddAttachment(
		&BotResponseAttachment{
			Color: "#00ff00",	
			Text: fmt.Sprintf("You roll %s", strings.Join(results, ", ")),
			Title: fmt.Sprintf("Roll %d %d-sided dice", qty, sides),
		},
	)
	return r, true
}

func NewPluginDice(b *Bot) *PluginDice {
	return &PluginDice{
		PluginBase: PluginBase{
			Bot: b,
		},
	}
}
