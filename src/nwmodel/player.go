package nwmodel

import (
	"fmt"
	"strconv"

	"nwmessage"

	"github.com/gorilla/websocket"
)

// TODO un export all but route
type playerID = int

var playerIDCount playerID

// Player ...
type Player struct {
	ID       playerID               `json:"id"`
	name     string                 `json:"name"`
	TeamName string                 `json:"team"`
	Route    *route                 `json:"route"`
	Socket   *websocket.Conn        `json:"-"`
	Outgoing chan nwmessage.Message `json:"-"`
	language string                 // current working language

	slotNum   int                 // currently attached to slotNum of current node
	dialogue  *nwmessage.Dialogue // this holds any dialogue the players in the middle of
	compiling bool                // this is used to block player action while submitted code is compiling
	ChatMode  bool                // track whether player is in chatmode or not (for use in lobby)
	inGame    bool                // is player in a game?

	EditorState string
	StdinState  string // stdin buffer for testing
	// termState terminalState
}

// player methods -------------------------------------------------------------------------------
// TODO this is in the wrong place
func NewPlayer(ws *websocket.Conn) *Player {
	ret := &Player{
		ID:       playerIDCount,
		name:     "",
		Socket:   ws,
		Outgoing: make(chan nwmessage.Message),
		slotNum:  -1,

		EditorState: "",
		StdinState:  "",
	}

	// log.Println("New player created, setting language...")
	playerIDCount++
	return ret
}

func (p *Player) stdinState(s string) {
	p.StdinState = s
	p.Outgoing <- nwmessage.StdinState(p.StdinState)
}

func (p *Player) editorState(s string) {
	p.EditorState = s
	p.Outgoing <- nwmessage.EditState(s)
}

func (p *Player) sendPrompt() {
	p.Outgoing <- nwmessage.PsPrompt(p.Prompt())
}

// Prompt should be generated by the ROOM the player is in...
func (p *Player) Prompt() string {
	if p.dialogue != nil {
		return ""
	}

	promptEndChar := ">"
	prompt := fmt.Sprintf("%s", p.GetName())

	if p.inGame {
		// if p.TeamName != "" {
		// 	prompt += fmt.Sprintf(":%s:", p.TeamName)
		// }
		if p.Route != nil {
			prompt += fmt.Sprintf("@n%d", p.Route.Endpoint.ID)
		}

		if p.slotNum != -1 {
			prompt += fmt.Sprintf(":s%d", p.slotNum)
		}
		prompt += fmt.Sprintf("[%s]", p.language)
	} else {
		prompt += "@lobby"
	}

	prompt += promptEndChar

	return prompt
}

// TODO refactor this, modify how slots are tracked, probably with IDs
func (p *Player) currentMachine() *machine {
	if p.Route == nil || p.slotNum < 0 || p.slotNum >= len(p.Route.Endpoint.Machines) {
		return nil
	}

	return p.Route.Endpoint.Machines[p.slotNum]
}

// GetName returns the players name if they have one, assigns one if they don't
func (p *Player) GetName() string {
	for p.name == "" {
		p.name = "player_" + strconv.Itoa(p.ID)
	}

	return p.name
}

func (p *Player) SetName(n string) {
	p.name = n
}

// hasTeam is deprecated I think TOD
func (p Player) hasTeam() bool {
	if p.TeamName == "" {
		return false
	}
	return true
}
