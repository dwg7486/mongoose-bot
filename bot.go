package main

import (
    "os"
    "os/signal"
    "fmt"
    "flag"
    "strings"
    "bufio"

    "github.com/bwmarrin/discordgo"
)

var (
    session *discordgo.Session

    OWNER_ID string

    /* temporary constants */
    GENERAL_CHANNEL string = "162620290487025674"

)

func handleOnReady(s *discordgo.Session, ev *discordgo.Ready) {
    s.UpdateStatus(0, "")
}

func handleMessageCreate(s *discordgo.Session, ev *discordgo.MessageCreate) {

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
        session.ChannelMessageSend(GENERAL_CHANNEL, cmdBody)

    case "tts":
        session.ChannelMessageSendTTS(GENERAL_CHANNEL, cmdBody)

    }
}


func main() {
    var (
        Token = flag.String("t", "", "Discord Auth Token")
        Owner = flag.String("o", "", "Bot Owner ID")
        err error
    )
    flag.Parse()

    if *Token == "" {
        fmt.Println("Auth Token is required")
    }

    if *Owner == "" {
        fmt.Println("Owner ID is required")
        return
    }

    fmt.Println("Creating Discord session")

    session, err = discordgo.New(*Token)

    if err != nil {
        fmt.Println("Error creating Discord session")
        return
    }

    session.AddHandler(handleOnReady)
    session.AddHandler(handleMessageCreate)

    session.Open()

    go acceptStdIn()

    fmt.Println("Session initialization finished")

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, os.Kill)
    <-quit
}
