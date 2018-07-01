package engine

import "strings"
import "bot/config"

type BotRequest struct {
	ChannelID string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	TeamDomain string `json:"team_domain"`
	TeamID string `json:"team_id"`
	PostID string `json:"post_id"`
	Text string `json:"text"`
	Timestamp int64 `json:"timestamp"`
	Token string `json:"token"`
	TriggerWord string `json:"trigger_word"`
	UserID string `json:"user_id"`
	UserName string `json:"user_name"`
}

func (br *BotRequest) CommandAndArgs(maxArgs int) (string, []string) {
	s := br.Text
	if br.TriggerWord != "" {
		s = s[len(br.TriggerWord):]
	}
	s = strings.Trim(s, " \t\n\r")
	if maxArgs == 0 {
		return strings.Fields(s)[0], []string(nil)
	}	
	fields := strings.SplitN( s, " ", maxArgs+1 )
	return fields[0], fields[1:]	
}

type BotResponse struct {
	UserName string `json:"username"`
	Text string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
	ResponseType string `json:"response_type"`
	Attachments []*BotResponseAttachment
}

func (br *BotResponse) AddAttachment(a *BotResponseAttachment) {
	if br.Attachments == nil {
		br.Attachments = make([]*BotResponseAttachment, 0)
	}
	br.Attachments = append(br.Attachments, a)
}

type BotResponseAttachment struct {
	Color string `json:"color,omitempty"`
	Text string `json:"text,omitempty"`
	Pretext string `json:"pretext,omitempty"`
	Fallback string `json:"fallback,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	AuthorIcon string `json:"author_icon,omitempty"`
	AuthorLink string `json:"author_link,omitempty"`
	Title string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Fields []*BotResponseAttachmentField `json:"fields,omitempty"`
}

type BotResponseAttachmentField struct {
	Short bool `json:"short"`
	Title string `json:"title"`
	Value string `json:"value"`
} 

type Bot struct {
	Config *config.Config
	Plugins []Plugin
}
