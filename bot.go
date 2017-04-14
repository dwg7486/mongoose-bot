package main

import (
    "os"
    "os/signal"
    "fmt"
    "flag"
    "strings"
    "bufio"
    "net/http"
    "net/url"
    "io/ioutil"
    "encoding/json"

    "github.com/bwmarrin/discordgo"

    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var (
    session *discordgo.Session

    OWNER_ID string

    /* temporary constants */
    GENERAL_CHANNEL   = "162620290487025674"
    DEV_CHANNEL       = "301146902777561088"
    NSFW_CHANNEL      = "215653449298083841"

    db *sql.DB

)

func handleOnReady(s *discordgo.Session, ready *discordgo.Ready) {
    s.UpdateStatus(0, "")
}

func handleMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
    if strings.HasPrefix(msg.Content, "https://instagram.com/p/") || strings.HasPrefix(msg.Content, "http://instagram.com/p/") || strings.HasPrefix(msg.Content, "https://www.instagram.com/p/") || strings.HasPrefix(msg.Content, "http://www.instagram.com/p/") {

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

            s.ChannelMessageSend(GENERAL_CHANNEL, "Let me fix that for you: ")
            fixedUrl := string(*respMap["message"])
            fixedUrl = strings.Replace(fixedUrl, "\\", "", -1)
            s.ChannelMessageSend(GENERAL_CHANNEL, fixedUrl)

        }
    }

    if strings.HasPrefix(msg.Content, "sqlite>") {
        splitIn := strings.SplitN(msg.Content, " ", 2)
        if len(splitIn) == 2 {
            sqlQuery := splitIn[1]
            result := QueryDb(sqlQuery)
            if len(result) > 0 {
                fmt.Println(s.ChannelMessageSend(msg.ChannelID, result))
            }
        }
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
        session.ChannelMessageSend(GENERAL_CHANNEL, cmdBody)

    case "tts":
        session.ChannelMessageSendTTS(GENERAL_CHANNEL, cmdBody)

    }
}

func InitSqlPrompt() {
    var err error
    db, err = sql.Open("sqlite3", "./db/mbot.db")
    if err != nil {
        panic(err)
    }
}

func QueryDb(query string) string {
    var result string = "No rows matched your query."
    rows, err := db.Query(query)
    if err != nil {
        result = "Database query failed."
    } else {
        var id int

        for rows.Next() {
            err = rows.Scan(&id, &result)
        }
    }
    return result
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

    fmt.Println("Session initialization finished")

    go acceptStdIn()
    InitSqlPrompt()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, os.Kill)
    <-quit
}