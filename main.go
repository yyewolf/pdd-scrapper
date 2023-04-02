package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/state"
	"github.com/diamondburned/ningen"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func init() {
	godotenv.Load()
}

var session *ningen.State
var targetID = discord.UserID(716390085896962058)

func main() {
	var token = os.Getenv("TOKEN")

	s, err := state.New(token)
	if err != nil {
		panic(errors.Wrap(err, "failed to create state"))
	}

	n, err := ningen.FromState(s)
	if err != nil {
		panic(errors.Wrap(err, "failed to create ningen"))
	}
	n.AddHandler(OnMessageCreate)
	session = n

	if err := n.Open(); err != nil {
		panic(errors.Wrap(err, "failed to open ningen"))
	}

	n.Gateway.GuildSubscribe(gateway.GuildSubscribeData{
		Typing:     true,
		Activities: true,
		GuildID:    discord.GuildID(716390832034414685),
		Channels: map[discord.ChannelID][][2]int{
			discord.ChannelID(784148997583890836): {
				{0, 99},
			},
		},
	})

	// Wait for CTRL+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Press Ctrl+C to stop the program...")
	<-done
}

func pad(n int, width int) string {
	return fmt.Sprintf("%0[2]*[1]d", n, width)
}

func OnMessageCreate(m *gateway.MessageCreateEvent) {
	// Check if the message comes from the target ID
	if m.Author.ID != targetID {
		return
	}

	// Check if the message contains "You caught a level"
	if !strings.Contains(m.Content, "You caught a level") {
		return
	}

	// Get the pokémon name
	pokemon := strings.Split(m.Content, "You caught a level ")[1]
	pokemon = strings.Join(strings.Split(pokemon, " ")[1:], " ")
	pokemon = strings.Split(pokemon, "!")[0]

	// Query the channel's messages from the state
	msgs, err := session.Store.Messages(m.ChannelID)
	if err != nil {
		panic(errors.Wrap(err, "failed to get messages from state"))
	}

	// Find the previous message that contains "A wild pokémon has appeared!"
	var lastMsg *discord.Message
	// for _, msg := range msgs {
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		if len(msg.Embeds) == 0 {
			continue
		}
		if msg.Embeds[0].Title == "A wild pokémon has appeared!" {
			lastMsg = &msg
		}
	}

	// If the previous message is not found, return
	if lastMsg == nil {
		return
	}

	// Get the pokémon's image URL
	var imageURL string
	for _, embed := range lastMsg.Embeds {
		if embed.Image != nil {
			imageURL = embed.Image.URL
		}
	}

	if imageURL == "" {
		return
	}

	// Save the pokémon's image in "out/{name}/000.png"
	// where {name} is the pokémon's name
	r, err := http.Get(imageURL)
	if err != nil {
		panic(errors.Wrap(err, "failed to get image"))
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(errors.Wrap(err, "failed to read body"))
	}

	r.Body.Close()

	// Create the directory if it doesn't exist
	if _, err := os.Stat("out/" + pokemon); os.IsNotExist(err) {
		os.MkdirAll("out/"+pokemon, os.ModePerm)
	}

	// Count how many images are in the directory
	var count int
	if _, err := os.Stat("out/" + pokemon); !os.IsNotExist(err) {
		files, err := os.ReadDir("out/" + pokemon)
		if err != nil {
			panic(errors.Wrap(err, "failed to read directory"))
		}
		count = len(files)
	}

	// Do not save if there's 10 images
	if count >= 10 {
		return
	}

	// Save the image
	f, err := os.Create("out/" + pokemon + "/" + pad(count, 3) + ".png")
	if err != nil {
		panic(errors.Wrap(err, "failed to create file"))
	}

	_, err = f.Write(body)
	if err != nil {
		panic(errors.Wrap(err, "failed to write file"))
	}

	fmt.Println("Saved " + pokemon + " #" + pad(count, 3) + " to out/" + pokemon + "/" + pad(count, 3) + ".png")
}
