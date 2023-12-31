package dc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
)

var (
	BotToken string
	// Discord server ID
	StransyyyBotChanneId string
)

// JSON Quote data
type QuoteData struct {
	Quotes []struct {
		Quote  string `json:"quote"`
		Author string `json:"author"`
	} `json:"quotes"`
}

type Record struct {
	RecordId int64
}

const prefix string = "!bot"

func Run(db *sql.DB) {
	if db == nil {
		log.Fatal("Error db is nil in Run")
		return
	}

	// Creates a new discord session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal("Error creating discord session", err)
	}

	// Add event handler for general messages
	discord.AddHandler(newMessage)

	// Open's and closes with the defer the session

	err = discord.Open()
	if err != nil {
		log.Fatal("Could not open session: ", err)
	}

	discord.SyncEvents = false

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"clockin":      func(s *discordgo.Session, i *discordgo.InteractionCreate) { ClockInResponse(s, i, db) },
		"clockout":     func(s *discordgo.Session, i *discordgo.InteractionCreate) { ClockOutResponse(s, i, db) },
		"lastInserted": func(s *discordgo.Session, i *discordgo.InteractionCreate) { lastInsertedResponse(s, i, db) },
	}

	discord.Identify.Intents = discordgo.IntentGuildMessages

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	var commands []*discordgo.ApplicationCommand

	commands = append(commands, clockinTimeCommand(), clockoutTimeCommand(), lastInsertedCommand())

	for _, c := range commands {
		_, cmderr := discord.ApplicationCommandCreate(os.Getenv("BOT_APP_ID"), "", c)

		if cmderr != nil {
			log.Fatal("This is the commands error at line 76", cmderr)
		}

		_, cmderr = discord.ApplicationCommandCreate(os.Getenv("BOT_APP_ID"), "", c)
	}

	defer discord.Close()

	fmt.Println("Bot running...")

	// Exit the session/program CTRL + C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func QuotesSend() []string {
	// Read the JSON file
	fileContent, err := os.ReadFile("/home/alan/src/golang-api-db/Sheet-Linker/dc/quotes.json")
	if err != nil {
		fmt.Println("Error reading the file:", err)
		return nil
	}

	// Create an instance of QuoteData
	var data QuoteData

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)

		return []string{}
	}

	var quotes []string

	for _, q := range data.Quotes {
		quotes = append(quotes, fmt.Sprintf("%s - %s", q.Quote, q.Author))
	}

	return quotes
}

// newMessage sends a new message. Does not reply to slash commands
func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignores bot own messages
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// Handles the mention of the user
	mencionString := "<@" + message.Author.ID + ">"

	// Respond to User messages using switch statementso we can answer a set of  predetermined messages
	switch {
	case strings.Contains(message.Content, "time"):
		discord.ChannelMessageSend(message.ChannelID, "I can provide that information")
	case strings.Contains(message.Content, "hola"):
		discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Hola %s", mencionString))
	}

	// let the user use !bot and the key word just for the bot to reply to that specific input
	args := strings.Split(message.Content, " ")
	if args[0] != prefix {
		return
	}

	// Access the quotes as a slice of strings
	quotes := QuotesSend()

	// Selects a random quote from the slice of strings of quotes
	selection := rand.Intn(len(quotes))

	// The bot prints a random quote using the !bot prefix
	if args[1] == "quotes" {
		discord.ChannelMessageSend(message.ChannelID, quotes[selection])
	}

}

// clockInTimeCommand on discord is the part where you use the slash command and shows a preview with the name of the command and the description of it
func clockinTimeCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "clockin",
		Description: "Run this command to clock in to work!",
	}

}

// clockoutTimeCommand is the slash command that displays on discord and shows the description of it
func clockoutTimeCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "clockout",
		Description: "Run this command to clockout from work, and send the data to the database",
	}

}

func getDataFromClockInCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "last-inserted",
		Description: "Get the last input of the database",
	}
}

// clockInEmbed will send an embed
func clockInEmbed() *discordgo.MessageEmbed {
	image := discordgo.MessageEmbedImage{
		URL: "https://pics.craiyon.com/2023-11-07/4db06060d78340a29c18a0436d9eaa56.webp",
	}

	// Embed content
	embed := discordgo.MessageEmbed{
		URL:         "https://vitalitysouth.com/",
		Title:       "Clock-In",
		Description: "",
		Color:       5763719,
		Image:       &image,
	}

	return &embed
}

func clockInResponseEmbed() *discordgo.MessageEmbed {
	image := discordgo.MessageEmbedImage{
		URL: "https://pics.craiyon.com/2023-09-10/25964e1a01134ba2b5497645ee40b85a.webp",
	}

	// Embed content
	embed := discordgo.MessageEmbed{
		URL:         "https://vitalitysouth.com/",
		Title:       "Clock-In",
		Description: "",
		Color:       5763719,
		Image:       &image,
	}

	return &embed
}

// clockOutEmbed is the set up for the clock-out embed
func clockOutEmbed() *discordgo.MessageEmbed {

	image := discordgo.MessageEmbedImage{
		URL: "https://cdn.dribbble.com/users/760079/screenshots/6979564/clockingout_4x.png",
	}

	// Embed content
	embed := &discordgo.MessageEmbed{
		URL:         "https://vitalitysouth.com/",
		Title:       "Clock-Out",
		Description: "",
		Color:       15548997,
		Image:       &image,
	}

	return embed
}

// ClockInResponse is the message the bot sends and the actions it takes whenever is being used
func ClockInResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, db *sql.DB) {
	if db == nil {
		log.Fatal("Error db is nil in clockinresponse")
		return
	}

	channelID := "1172648319940558970"
	_, err := session.ChannelMessageSend(channelID, "Stransyyy bot esta siendo usado...")
	if err != nil {
		log.Fatal("Is the error here?", err)
	}

	// Check if interaction or interaction.Interaction is nil
	if interaction == nil || interaction.Interaction == nil {
		log.Println("Invalid interaction object")
		return
	}

	userID := func() string {
		if interaction.User != nil {
			return interaction.User.ID
		}

		if interaction.Member != nil {
			return interaction.Member.User.ID
		}

		return ""
	}()

	// Respond to the slash command interaction with a deferred response
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "You have clocked-in succesfully!",
		},
	})
	if err != nil {
		log.Println("Error responding to interaction:", err)
		return
	}

	//FollowUp message after you use the command
	mensaje := &discordgo.WebhookParams{
		Content: "You have Clocked-in successfuly",
	}

	_, err = session.FollowupMessageCreate(interaction.Interaction, false, mensaje)
	if err != nil {
		return
	}

	//Stores the data into the database
	messageId := interaction.ID
	authorId := interaction.Member.User.ID
	msgContent := "/clockin"
	dateSent, err := discordgo.SnowflakeTimestamp(interaction.ID)
	serverId := interaction.GuildID
	channelId := interaction.ChannelID

	if err != nil {
		log.Println("dateSent snowflake error ", err)
		dateSent = time.Now()
	}

	_, dberr := messagesDataBaseClockInHandler(db, serverId, channelId, messageId, authorId, msgContent, dateSent.Local())

	if dberr != nil {
		log.Fatal("messageDatabase interaction message not working:", err)
	}

	// Create a DM channel with the user who used the command
	dmChannel, err := session.UserChannelCreate(userID)
	if err != nil {
		log.Println("Error creating DM channel:", err)
		return
	}

	// Send a message with an embed to the user in the DM channel
	if dmChannel != nil && dmChannel.ID != "" {
		_, dmErr := session.ChannelMessageSendEmbed(dmChannel.ID, clockInEmbed())

		if dmErr != nil {
			log.Println("Error sending DM with embed:", dmErr)
		}
	} else {
		log.Println("Invalid DM channel")
	}
}

// clockOutResponse will send an embed as a response to the slash command call
func ClockOutResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, db *sql.DB) {
	if db == nil {
		log.Fatal("Error db is nil in clockoutresponse")
		return
	}

	channelID := "1172648319940558970"
	_, ok := session.ChannelMessageSend(channelID, "Stransyyy bot esta siendo usado...")
	if ok != nil {
		log.Fatal("Is the error here?", ok)
	}

	// Check if interaction or interaction.Interaction is nil
	if interaction == nil || interaction.Interaction == nil {
		log.Println("Invalid interaction object")
		return
	}

	userID := func() string {
		if interaction.User != nil {
			return interaction.User.ID
		}

		if interaction.Member != nil {
			return interaction.Member.User.ID
		}

		return ""
	}()

	// Respond to the slash command interaction with a deferred response
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "You have clocked-Out succesfully!",
		},
	})
	if err != nil {
		log.Println("Error responding to interaction:", err)
		return
	}

	//FollowUp message after you use the command
	mensaje := &discordgo.WebhookParams{
		Content: "You have Clocked-Out successfuly, you can now rest!",
	}

	_, err = session.FollowupMessageCreate(interaction.Interaction, false, mensaje)
	if err != nil {
		return
	}

	// Create a DM channel with the user who used the command
	dmChannel, err := session.UserChannelCreate(userID)
	if err != nil {
		log.Println("Error creating DM channel:", err)
		return
	}

	// The database interaction
	messageId := interaction.ID
	authorId := interaction.Member.User.ID
	msgContent := "/clockout"
	now := time.Now().Unix()

	prueba := time.Unix(now, 0)

	///

	test := prueba.Format("2006-01-02 15:04:05")

	fmt.Println(test)

	///

	serverId := interaction.GuildID
	channelId := interaction.ChannelID

	err = messagesDataBaseClockOutHandler(db, serverId, channelId, messageId, authorId, msgContent, now)

	// Send a message with an embed to the user in the DM channel (this will be something else)
	if dmChannel != nil && dmChannel.ID != "" {
		_, dmErr := session.ChannelMessageSendEmbed(dmChannel.ID, clockOutEmbed())

		if dmErr != nil {
			log.Println("Error sending DM with embed:", dmErr)
		}
	} else {
		log.Println("Invalid DM channel")
	}
}

// messageDataBaseClockInHandler stores the data using the query that needs to be sent everytime the slash command is executed in discord
func messagesDataBaseClockInHandler(db *sql.DB, serverId, channelId, messageId, authorId, content string, dateSent time.Time) (error, sql.Result) {
	query := "INSERT INTO messages (message_id, guild_id, channel_id, author_id, message_content, time_sent) VALUES (?, ?, ?, ?, ?, ?)"

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Couldn't initialize the transaction:", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, errs := tx.Exec(query, messageId, serverId, channelId, authorId, content, dateSent)
	if errs != nil {
		log.Fatal("Couldn't insert into the database:", errs)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Couldn't commit the transaction: %v", err), nil
	}

	return nil, res
}

// messageDataBaseClockOutHandler stores the data using the query that needs to be sent everytime the slash command is executed in discord
func messagesDataBaseClockOutHandler(db *sql.DB, serverId, channelId, messageId, authorId, content string, dateSent int64) error {
	query := "INSERT INTO messages (message_id, guild_id, channel_id, author_id, message_content, time_sent) VALUES (?, ?, ?, ?, ?, ?)"

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Couldn't initialize the transaction:", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, errs := tx.Exec(query, messageId, serverId, channelId, authorId, content, dateSent)
	if errs != nil {
		log.Fatal("Couldn't insert into the database:", errs)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Couldn't commit the transaction: %v", err)
	}

	return nil
}
