package engine

func GetPlugins(b *Bot) []Plugin {

	return []Plugin{
		NewPluginFeed(b),
		NewPluginDice(b),
		NewPluginGem(b),
	}
	
}