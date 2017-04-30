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
	"strconv"
	"bytes"
)

var (
	session *discordgo.Session

	OWNER_ID           string
	GENERAL_CHANNEL_ID string = "162620290487025674"
)

func HandleOnReady(s *discordgo.Session, ready *discordgo.Ready) {
	s.UpdateStatus(0, "")
}

func HandleMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
	isInstagramLink, _ := regexp.MatchString("https?://w{0,3}\\.?instagram\\.com/p/", msg.Content)
	if isInstagramLink {
		go FixInstagramLink(s, msg)
	}

	// ------------------------------------------------------------------------
	if msg.ChannelID == "267497667448274945" {
		if !(msg.Author.ID == "162536493137657856" || msg.Author.ID == "163845105063559169") {
			go s.ChannelMessageDelete("267497667448274945", msg.ID)
		}
	}
	// ------------------------------------------------------------------------

	if strings.HasPrefix(msg.Content, "!ev") || strings.HasPrefix(msg.Content, "!event") {

		// Only work for me
		if msg.Author.ID != "162689822404640771" {
			s.ChannelMessageSend(msg.ChannelID, "Sorry, I can only do that for Mongoose for now.")
			return
		}
		// ----------------
		splitIn := strings.SplitN(msg.Content, " ", 2)

		splitCmd := strings.SplitN(splitIn[1], " ", 2)
		cmdKey := splitCmd[0]
		var cmdBody string = ""
		if len(splitCmd) == 2 {
			cmdBody = splitCmd[1]
		}

		switch cmdKey {
		case "create":
			splitBody := strings.Split(cmdBody, "|")
			if len(splitBody) != 5 {
				s.ChannelMessageSend(msg.ChannelID, "Usage: !event create name|description|location|date|time")
				break
			}

			name := splitBody[0]
			description := splitBody[1]
			location := splitBody[2]
			date := splitBody[3]
			time := splitBody[4]

			event, err := CreateEvent(
				name,
				description,
				location,
				date,
				time,
				msg.Author.Username,
				msg.Author.ID)

			channel, channelErr := s.UserChannelCreate(msg.Author.ID)
			PanicIf(channelErr)
			if err == nil {
				s.ChannelMessageSend(channel.ID,
					"**Created event:** "+name+"\n"+
						"**Description:** "+description+"\n"+
						"**When:** "+event.date+" at "+event.time+"\n"+
						"**Where:** "+location+"\n"+
						"Your event ID is "+strconv.FormatInt(event.id, 10)+".\n"+
						"Remember this ID if you wish to make changes to your event.")
			} else {
				s.ChannelMessageSend(channel.ID,
					"Event creation failed.  Please make sure you are using the command correctly.")
			}

		case "info":
			eventSearch := cmdBody
			events, err := RetrieveEvent(eventSearch)
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				break
			}
			if len(events) > 1 {
				var buffer bytes.Buffer
				for _, event := range events {
					buffer.WriteString(
						"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
							" at " + event.time + "`\n")
				}
				s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
					buffer.String()+"Select one by its ID with !event info <ID> for more information.")

			} else if len(events) == 1 {
				event := events[0]
				rsvps, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
				if err != nil {
					s.ChannelMessageSend(msg.ChannelID, err.Error())
					break
				}
				var buffer bytes.Buffer
				for i, rsvp := range rsvps {
					buffer.WriteString("__"+rsvp.username+"__: "+rsvp.status)
					if i < (len(rsvps)-1) {
						buffer.WriteString("    ")
					}
				}
				s.ChannelMessageSend(msg.ChannelID, event.String())
			} else {
				s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
			}

		case "cancel":
			eventSearch := cmdBody
			events, err := RetrieveEvent(eventSearch)
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
			} else {
				if len(events) > 1 {
					var buffer bytes.Buffer
					for _, event := range events {
						buffer.WriteString(
							"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
								" at " + event.time + "`\n")
					}
					s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
						buffer.String()+"Select one by its ID with !event info <ID> for more information.")

				} else if len(events) == 1 {
					event := events[0]
					if msg.Author.ID == event.creatorID {
						err = CancelEvent(strconv.FormatInt(event.id, 10))
						if err != nil {
							s.ChannelMessageSend(msg.ChannelID, err.Error())
						} else {
							s.ChannelMessageSend(msg.ChannelID, "Event cancelled.")
						}
					} else {
						s.ChannelMessageSend(msg.ChannelID, "You can't cancel an event you didn't create.")
					}
				} else {
					s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
				}
			}

		case "edit":
			splitBody := strings.Split(cmdBody, "|")
			if len(splitBody) != 3 {
				s.ChannelMessageSend(msg.ChannelID, "Usage: !event edit eventID|fieldName|newValue\n"+
					"Editable field names are desc[ription], loc[ation], date, and time.  Event name is not editable.")
				break
			}

			eventSearch := splitBody[0]
			column := splitBody[1]
			newValue := splitBody[2]

			events, err := RetrieveEvent(eventSearch)
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				break
			}

			if len(events) > 1 {
				var buffer bytes.Buffer
				for _, event := range events {
					buffer.WriteString(
						"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
							" at " + event.time + "`\n")
				}
				s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
					buffer.String()+"Select one by its ID with !event info <ID> for more information.")

			} else if len(events) == 1 {
				event := events[0]
				id := strconv.FormatInt(event.id, 10)
				switch strings.ToLower(column) {
				case "description", "desc":
					UpdateEventDescription(id, newValue)
				case "location", "loc":
					UpdateEventLocation(id, newValue)
				case "date":
					UpdateEventDate(id, newValue)
				case "time":
					UpdateEventTime(id, newValue)
				default:
					s.ChannelMessageSend(msg.ChannelID,
						"Editable field names are desc[ription], loc[ation], date, and time.  Event name is not editable.")
				}

			} else {
				s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
			}

			/*
				eventID := cmdBody
				getEvent := `
					SELECT name,desc,datetime,location FROM events
					WHERE id = ?
					`

				result := db.QueryRow(getEvent, eventID)

				if result != nil {
					var (
						name     string
						desc     string
						datetime string
						location string
					)

					err := result.Scan(&name, &desc, &datetime, &location)
					if err == nil {

						findRSVPs := `
							SELECT personid,status FROM rsvps
							WHERE eventid = ?
						`
						rows, _ := db.Query(findRSVPs, eventID)
						defer rows.Close()

						var rsvps bytes.Buffer
						for rows.Next() {
							var (
								userID string
								user   *discordgo.User
								status string
							)
							err := rows.Scan(&userID, &status)
							if err != nil {
								return
							}

							user, _ = s.User(userID)

							rsvps.WriteString(user.Username + ": " + status + "\n")
						}

						s.ChannelMessageSend(msg.ChannelID,
							"__**"+name+"**__\n"+
								"__When:__ "+datetime+"\n"+
								"__Where:__ "+location+"\n"+
								"*"+desc+"*\n\n"+
								rsvps.String())
					} else {
						s.ChannelMessageSend(msg.ChannelID,
							"No events matched your query. :slight_frown:")
					}
				}*/
		case "rsvp":
			splitBody := strings.Split(cmdBody, "|")
			if len(splitBody) != 2 {
				s.ChannelMessageSend(msg.ChannelID, "Usage: !event rsvp eventID|choice")
				break
			}

			choice := splitBody[1]
			// Validate RSVP choice
			switch strings.ToLower(choice) {
			case "g", "going":
				choice = "Going"

			case "m", "maybe":
				choice = "Maybe"

			case "n", "not going":
				choice = "Not going"

			default:
				s.ChannelMessageSend(msg.ChannelID,
					"Valid RSVP choices: G[oing], M[aybe], N[ot going]")
				return
			}

			eventSearch := splitBody[0]
			events, err := RetrieveEvent(eventSearch)
			if err != nil {
				s.ChannelMessageSend(msg.ChannelID, err.Error())
				break
			}

			if len(events) > 1 {
				var buffer bytes.Buffer
				for _, event := range events {
					buffer.WriteString(
						"ID: " + strconv.FormatInt(event.id, 10) + " `**" + event.name + "** on " + event.date +
							" at " + event.time + "`\n")
				}
				s.ChannelMessageSend(msg.ChannelID, "Your query matched the following events:\n"+
					buffer.String()+ "Select one by its ID with !event info <ID> for more information.")

			} else if len(events) == 1 {
				event := events[0]
				_, err := RetrieveRSVPs(strconv.FormatInt(event.id, 10))
				if err != nil {
					s.ChannelMessageSend(msg.ChannelID, err.Error())
					break
				}

			} else {
				s.ChannelMessageSend(msg.ChannelID, "No events found.  Try a different search.")
			}

			/*
			findExistingRSVP := `
		SELECT id FROM rsvps
		WHERE eventid = ? AND personid = ?
		`
			db.QueryRow(findExistingRSVP, eventID, msg.Author.ID).Scan(&existingID)

			if existingID > 0 {
				updateExistingRSVP := `
				UPDATE rsvps
				SET status = ?
				WHERE id = ?
				`
				db.Exec(updateExistingRSVP, choice, existingID)

				s.ChannelMessageSend(msg.ChannelID,
					"Updated RSVP for "+msg.Author.Username)
				return
			} else {
				makeRSVP := `
			INSERT INTO rsvps(
				eventid,
				personid,
				status
			) VALUES (?, ?, ?)
			`

				_, err := db.Exec(makeRSVP, eventID, msg.Author.ID, choice)
				if err != nil {
					fmt.Println("Error creating RSVP")
					panic(err)
				}

				s.ChannelMessageSend(msg.ChannelID,
					"Submitted RSVP for "+msg.Author.Username+".")
			}
			*/

		case "help":
			s.ChannelMessageSend(msg.ChannelID,
				"__Discord Event Planner created by Mongoose__"+"\n```"+
					"**Create event:** !event create name|description|location|date|time"+"\n"+
					"**Edit event:**   !event edit eventID|fieldName|newValue"+"\n"+
					"**Cancel event**  !event cancel eventID"+"\n"+
					"**Show event**    !event info eventID  OR  !event info eventName"+"\n"+
					"**RSVP**          !event rsvp eventID|choice OR  !rsvp eventName|choice"+"\n"+
					"**RSVP choices:** G[oing], M[aybe], N[ot going]"+"```")
		}
	}
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

	//InitSqlConnection()

	fmt.Println("Session initialization finished")

	go acceptStdIn()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
