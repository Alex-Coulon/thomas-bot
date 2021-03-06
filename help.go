package main

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

// map[category][]map[command]description
var helpData = map[string]map[string]string{}

func init() {
	registerCommand("help", "info", "Lijst van alle commandus (u bent hier)", sayHelp)
}

func sayHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := NewEmbed()
	embed.SetTitle("Help")

	categories := make([]string, 0, len(helpData))
	for k := range helpData {
		categories = append(categories, k)
	}
	sort.Strings(categories)
	for _, categoryname := range categories {
		commandoKeys := make([]string, 0, len(helpData[categoryname]))
		for k := range helpData[categoryname] {
			commandoKeys = append(commandoKeys, k)
		}
		sort.Strings(commandoKeys)

		out := ""
		for _, commandoName := range commandoKeys {
			out += fmt.Sprintf("* `%s`: %s\n", commandoName, helpData[categoryname][commandoName])
		}
		embed.AddField(categoryname, out)
	}
	s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
}
