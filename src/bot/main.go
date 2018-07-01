package main

import "bot/config"
import "bot/engine"
import "log"
import "os"

func main() {
	bots := make([]*engine.Bot, 0)
	args := os.Args[1:]
	for _, cfgName := range args {
		cfg, err := config.Load(cfgName)
		if err != nil {
			log.Printf("Skipping config %s: %v", cfgName, err)
			continue		
		}
		bot, err := engine.New(cfg)
		if err != nil {
			log.Printf("Failed to start bot with config %s: %v", cfgName, err)
			continue
		}
		bots = append(bots, bot)
	}
	if len(bots) == 0 {
		log.Println("Terminating as zero bots are running...")
		return
	}
	select{}
}