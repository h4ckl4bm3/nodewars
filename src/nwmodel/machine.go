package nwmodel

import (
	"math/rand"
	"sync"
)

type machine struct {
	sync.Mutex
	// accepts   challengeCriteria
	challenge Challenge
	// Type      string `json:"type"`
	Powered  bool   `json:"powered"`
	builder  string // `json:"creator"`
	TeamName string `json:"team"`
	// solution  string
	language  string // `json:"languageId"`
	Health    int    `json:"health"`
	MaxHealth int    `json:"maxHealth"`
}

type feature struct {
	Type string `json:"type"` // type of feature
	machine
}

type challengeCriteria struct {
	IDs        []int64  // list of acceptable challenge ids
	Tags       []string // acceptable categories of challenge
	Difficulty [][]int  // acceptable difficulties, [5] = level five, [3,5] = 3,4, or 5
}

// init methods

func newMachine() *machine {
	return &machine{Powered: true}
}

func newFeature() *feature {
	return &feature{
		machine: machine{Powered: true},
	}
}

// machine methods -------------------------------------------------------------------------
func (m *machine) resetChallenge() {
	m.challenge = getRandomChallenge()
}

func (m *machine) isNeutral() bool {
	if m.TeamName == "" {
		return true
	}
	return false
}

func (m *machine) belongsTo(teamName string) bool {
	if m.TeamName == teamName {
		return true
	}
	return false
}

func (m *machine) reset() {
	m.builder = ""
	m.TeamName = ""
	m.language = ""
	m.Powered = true

	m.loadChallenge()
	m.Health = 0
	m.MaxHealth = len(m.challenge.Cases)
}

func (m *machine) claim(p *Player, r ExecutionResult) {
	m.builder = p.name
	m.TeamName = p.TeamName
	m.language = p.language
	// m.Powered = true

	m.Health = r.passed()
	// m.MaxHealth = len(r.Graded)
}

// dummyClaim is used to claim a machine for a player without an execution result
func (m *machine) dummyClaim(p *Player, str string) {
	m.builder = p.name
	m.TeamName = p.TeamName
	m.language = p.language
	// m.Powered = true

	switch str {
	case "FULL":
		m.Health = m.MaxHealth
	case "RAND":
		m.Health = rand.Intn(m.MaxHealth) + 1
	case "MIN":
		m.Health = 1
	}
}

// loadChallenge should use m.accepts to get a challenge matching criteria TODO
func (m *machine) loadChallenge() {
	m.challenge = getRandomChallenge()
}
