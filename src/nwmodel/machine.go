package nwmodel

import (
	"feature"
	"math/rand"
	"sync"
)

type machine struct {
	sync.Mutex
	// accepts   challengeCriteria // store what challenges can fill this machine
	challenge Challenge
	// Type      string `json:"type"`
	Powered  bool   `json:"powered"`
	builder  string // `json:"creator"`
	TeamName string `json:"owner"`

	// solution  string // store solution used to pass. could be useful for later mechanics
	Type feature.Type `json:"type"` // NA for non-features, none or other feature.Type for features

	language  string // `json:"languageId"`
	Health    int    `json:"health"`
	MaxHealth int    `json:"maxHealth"`
}

// type feature struct {
// 	Type featureType `json:"type"` // type of feature
// 	machine
// }

type challengeCriteria struct {
	IDs        []int64  // list of acceptable challenge ids
	Tags       []string // acceptable categories of challenge
	Difficulty [][]int  // acceptable difficulties, [5] = level five, [3,5] = 3,4, or 5
}

// init methods

func newMachine() *machine {
	return &machine{
		Powered: true,
	}
}

func newFeature() *machine {
	return &machine{
		Type:    feature.None,
		Powered: true,
	}
}

// machine methods -------------------------------------------------------------------------

// resetChallenge should use m.accepts to get a challenge matching criteria TODO
func (m *machine) resetChallenge() {
	m.challenge = getRandomChallenge()
	m.MaxHealth = len(m.challenge.Cases)
}

func (m *machine) isNeutral() bool {
	if m.TeamName == "" {
		return true
	}
	return false
}

func (m *machine) isFeature() bool {
	// fmt.Printf("machine Type: %v", m.Type)
	// fmt.Printf("Feature NA: %v", feature.NA)
	// fmt.Printf("Equal: %v", m.Type == feature.NA)
	if m.Type == nil {
		return false
	}
	return true
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

	// if m.Type != nil { // reset feature type?
	// 	m.Type = feature.None
	// }

	m.Health = 0
	m.resetChallenge()
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
func (m *machine) dummyClaim(teamName string, str string) {
	// m.builder = p.name
	m.TeamName = teamName
	m.language = "python" // TODO find ore elegent solution for this
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

// func (m *machine) loadChallenge() {
// 	m.challenge = getRandomChallenge()

// 	fmt.Printf("Loaded challenge, %v\n", m.challenge)
// }
