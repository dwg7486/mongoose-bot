package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"

	"regexp"
)

var (
	session *discordgo.Session

	OWNER_ID           string
	GENERAL_CHANNEL_ID string = "162620290487025674"

	banned map[string]bool = make(map[string]bool)
)

func HandleOnReady(s *discordgo.Session, ready *discordgo.Ready) {
	s.UpdateStatus(0, "")
}

func HandleMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
	isInstagramLink, _ := regexp.MatchString("https?://w{0,3}\\.?instagram\\.com/p/", msg.Content)
	if isInstagramLink {
		go FixInstagramLink(s, msg)
	}

	if strings.HasPrefix(msg.Content, "!ev") || strings.HasPrefix(msg.Content, "!event") {
		go HandleCommandInput(s, msg)
	}


	if strings.Compare(msg.Content, "I am a filthy reposter.") == 0 {
		banned[msg.Author.ID] = false
	}
	if isBanned, _ := banned[msg.Author.ID]; isBanned {
		s.ChannelMessageDelete(msg.ChannelID, msg.ID)
	}
	if strings.HasPrefix(msg.Content, "http") {
		isRepost, _ := DetectRepost(msg.Content)
		if isRepost {
			s.ChannelMessageSend(msg.ChannelID, "Repost. You have been banned from posting until you say `I am a filthy reposter.`")
			banned[msg.Author.ID] = true
		}
	}

	RecordMessage(msg.Author.ID, msg.Content)
}

func FixInstagramLink(s *discordgo.Session, msg *discordgo.MessageCreate) {
	resp, err := http.PostForm("http://www.igeturl.com/get.php", url.Values{"url": {msg.Content}})

	if err != nil {
		fmt.Println("Error getting converted link")
		return
	}
	defer resp.Body.Close()

	body, err1 := ioutil.ReadAll(resp.Body)

	if err1 != nil {
		fmt.Println("Error reading response body")
		return
	}

	var respMap map[string]*json.RawMessage
	json.Unmarshal([]byte(body), &respMap)

	if string(*respMap["success"]) == "true" {
		s.ChannelMessageSend(msg.ChannelID, "Let me fix that for you: ")
		fixedUrl := string(*respMap["message"])
		fixedUrl = strings.Replace(fixedUrl, "\\", "", -1)
		fixedUrl = strings.Replace(fixedUrl, "\"", "", -1)
		s.ChannelMessageSend(msg.ChannelID, fixedUrl)
	}
}

func acceptStdIn() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, _ := reader.ReadString('\n')
		parseCommand(text)
	}
}

func parseCommand(input string) {
	splitIn := strings.SplitN(input, " ", 2)
	cmdKey := splitIn[0]
	cmdBody := splitIn[1]

	switch cmdKey {

	case "say":
		session.ChannelMessageSend(GENERAL_CHANNEL_ID, cmdBody)

	case "tts":
		session.ChannelMessageSendTTS(GENERAL_CHANNEL_ID, cmdBody)

	}

}

func main() {
	var (
		Token = flag.String("t", "", "Discord Auth Token")
		Owner = flag.String("o", "", "Bot Owner ID")
		err   error
	)
	flag.Parse()

	if *Token == "" {
		fmt.Println("Auth Token is required")
	}

	if *Owner == "" {
		fmt.Println("Owner ID is required")
		return
	}

	OWNER_ID = *Owner

	fmt.Println("Creating Discord session")

	session, err = discordgo.New(*Token)

	if err != nil {
		fmt.Println("Error creating Discord session")
		return
	}

	session.AddHandler(HandleOnReady)
	session.AddHandler(HandleMessageCreate)

	session.Open()

	fmt.Println("Session initialization finished")

	go acceptStdIn()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
