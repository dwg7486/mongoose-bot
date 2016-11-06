package main

import (
    "os"
    "os/signal"
    "fmt"
    "flag"

    "github.com/bwmarrin/discordgo"
)

func handleOnReady(session *discordgo.Session, event *discordgo.Ready) {
    session.UpdateStatus(0, "")
}

func main() {
    var (
        Token = flag.String("t", "", "Discord Auth Token")
        Owner = flag.String("o", "", "Bot Owner ID")
    )
    flag.Parse()

    fmt.Println(*Owner)

    fmt.Println("Creating Discord session")

    session, err := discordgo.New(*Token)

    if err != nil {
        fmt.Println("Error creating Discord session")
        return
    }

    session.AddHandler(handleOnReady)

    session.Open()

    fmt.Println("Session initialization finished")

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, os.Kill)
    <-quit
}
