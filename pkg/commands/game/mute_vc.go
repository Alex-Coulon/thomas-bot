package game

import (
	"fmt"
	"log"
	"strings"

	"github.com/itfactory-tm/thomas-bot/pkg/embed"

	"github.com/itfactory-tm/thomas-bot/pkg/command"

	"github.com/bwmarrin/discordgo"
	"github.com/itfactory-tm/thomas-bot/pkg/sudo"
)

// adduser contains the bob!adduser and bob!remuser command
type MuteCommand struct{}

// NewMuteCommand gives a new MuteCommand
func NewMuteCommand() *MuteCommand {
	return &MuteCommand{}
}

// Register registers the handlers
func (m *MuteCommand) Register(registry command.Registry, server command.Server) {
	registry.RegisterMessageCreateHandler("mutevc", m.mutevc)
	registry.RegisterMessageReactionAddHandler(m.handleMuteReaction)
}

func (m *MuteCommand) mutevc(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if !sudo.IsItfAdmin(msg.Author.ID) {
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("%s is not in the sudoers file. This incident will be reported.", msg.Author.ID))
		return
	}

	// Get the guild status
	g, err := s.State.Guild(msg.GuildID)
	if err != nil {
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Error getting guild: %v", err))
		return
	}

	channelId := strings.TrimPrefix(msg.Content, "bob!mutevc")
	if len(channelId) > 1 {
		channelId = strings.TrimPrefix(channelId, " ")
		//ChannelID specified, check if channel exists
		for _, channel := range g.Channels {
			if channel.ID == channelId && channel.Type == 2 {
				embedMsg, _ := s.ChannelMessageSendEmbed(msg.ChannelID, m.muteMenu(channel.ID, s).MessageEmbed)
				s.MessageReactionAdd(embedMsg.ChannelID, embedMsg.ID, "🔈")
				s.MessageReactionAdd(embedMsg.ChannelID, embedMsg.ID, "🔇")
				return
			}
		}
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("No voice channel found with id: %v", channelId))
		return
	} else {
		//ChannelID NOT specified, joining author channel
		for _, user := range g.VoiceStates {
			if user.UserID == msg.Author.ID {
				embedMsg, _ := s.ChannelMessageSendEmbed(msg.ChannelID, m.muteMenu(user.ChannelID, s).MessageEmbed)
				s.MessageReactionAdd(embedMsg.ChannelID, embedMsg.ID, "🔈")
				s.MessageReactionAdd(embedMsg.ChannelID, embedMsg.ID, "🔇")
				return
			}
		}
	}
	s.ChannelMessageSend(msg.ChannelID, "Please specify a voice channel by channelID or join a voice channel.")
	return
}

// Info return the commands in this package
func (m *MuteCommand) Info() []command.Command {
	return []command.Command{
		command.Command{
			Name:        "mutevc",
			Category:    command.CategoryModeratie,
			Description: "Mute everyone in a voice chat (ITF Game admin only)",
			Hidden:      false,
		},
	}
}

func (m *MuteCommand) muteMenu(id string, s *discordgo.Session) *embed.Embed {
	channel, _ := s.Channel(id)
	embed := embed.NewEmbed()
	embed.SetTitle("Mute")
	embed.AddField("Channel", channel.Name)
	embed.AddField("ChannelID", channel.ID)
	embed.InlineAllFields()
	embed.SetColor(3066993)
	embed.AddField("Status", "Unmuted!")
	return embed
}

func (m *MuteCommand) handleMuteReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Println("Cannot get message of reaction", r.ChannelID)
		return
	}

	if message.Author.ID != s.State.User.ID {
		return // not the bot user
	}

	if !sudo.IsItfAdmin(r.UserID) {
		return //Is not an Itf Admin
	}

	if len(message.Embeds) < 1 {
		return // not the mute message
	}

	if message.Embeds[0].Title != "Mute" {
		return //Not the mute message
	}

	// Get the guild status
	g, err := s.State.Guild(r.GuildID)
	if err != nil {
		s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("Error getting guild: %v", err))
		return
	}

	voiceChannel := message.Embeds[0].Fields[1].Value
	embed := m.muteMenu(voiceChannel, s)

	if r.Emoji.MessageFormat() == "🔈" {
		//Unmute everyone
		for _, user := range g.VoiceStates {
			if user.ChannelID == voiceChannel {
				s.GuildMemberMute(user.GuildID, user.UserID, false)
			}
		}
	}

	if r.Emoji.MessageFormat() == "🔇" {
		//mute everyone
		for _, user := range g.VoiceStates {
			if user.ChannelID == voiceChannel {
				s.GuildMemberMute(user.GuildID, user.UserID, true)
			}
		}
		embed.SetColor(15158332)
		embed.AddField("Status", "Muted!")
	}

	s.ChannelMessageEditEmbed(r.ChannelID, r.MessageID, embed.MessageEmbed)
}
