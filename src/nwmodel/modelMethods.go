package nwmodel

import (
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Initialization methods ------------------------------------------------------------------

func newModuleBy(p *Player) module {
	id := moduleIDCount
	moduleIDCount++

	return module{
		ID:         id,
		TestID:     0,
		LanguageID: 0,
		Builder:    p,
	}
}

// func newGameState() *gameState {
// 	// make a list of all team names
// 	teams := make([]string, 0)
// 	for name := range gm.Teams {
// 		teams = append(teams, name)
// 	}

// 	players := make([]*Player, 0)
// 	for player := range gm.Players {
// 		// Only add players to the stat object that have names and teams
// 		if player.hasName() && player.hasTeam() {
// 			players = append(players, player)
// 		}
// 	}

// 	return &gameState{
// 		Map:           *gm.Map,
// 		Teams:         teams,
// 		Players:       players,
// 		CurrentEvents: make([]gameEvent, 0),
// 	}
// }

// NewTeam creates a new team with color/name color
func NewTeam(n teamName) *team {
	return &team{n, make(map[*Player]bool), 2}
}

// NewNode ...
func NewNode() *node {
	id := nodeIDCount
	nodeIDCount++

	connections := make([]int, 0)
	modules := make(map[modID]module)

	return &node{
		ID:          id,
		Connections: connections,
		Size:        3,
		Modules:     modules,
		// Traffic:          make([]*Player, 0),
		// POE:              make([]*Player, 0),
		// ConnectedPlayers: make([]*Player, 0),
	}
}

// // hiLo is a helper function that lets new edge sort node pairs for its ID scheme
// func hiLo(a, b nodeID) (nodeID, nodeID) {
// 	if a > b {
// 		return a, b
// 	}
// 	return b, a
// }

// func newEdge(s, t nodeID) *edge {
// 	hi, lo := hiLo(s, t)

// 	hiStr := strconv.Itoa(hi)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	loStr := strconv.Itoa(lo)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }

// 	id := hiStr + "e" + loStr
// 	// id := edgeCount
// 	// edgeCount++

// 	return &edge{id, s, t, make([]*Player, 0)}
// }

func newNodeMap() nodeMap {
	return nodeMap{make([]*node, 0)}
}

// Instantiation with values ------------------------------------------------------------------

// NewDefaultModel Generic game model
func NewDefaultModel() *GameModel {
	m := newDefaultMap()
	t := makeDummyTeams()
	p := make(map[playerID]*Player)
	r := make(map[playerID]*route)
	poes := make(map[playerID]*node)

	return &GameModel{
		Map:     m,
		Teams:   t,
		Players: p,
		Routes:  r,
		POEs:    poes,
	}
}

func makeDummyTeams() map[teamName]*team {
	teams := make(map[teamName]*team)
	teams["red"] = NewTeam("red")
	teams["blue"] = NewTeam("blue")
	return teams
}

func newDefaultMap() *nodeMap {
	newMap := newNodeMap()

	NODECOUNT := 12

	for i := 0; i < NODECOUNT; i++ {
		//Make new nodes
		newMap.addNodes(NewNode())
	}

	for i := 0; i < NODECOUNT; i++ {
		//Make new edges
		targ1, targ2 := -1, -1

		if i < NODECOUNT-3 {
			targ1 = i + 3
		}

		if i%3 < 2 {
			targ2 = i + 1
		}

		if targ1 != -1 {
			newMap.connectNodes(i, targ1)
		}

		if targ2 != -1 {
			newMap.connectNodes(i, targ2)
		}
	}

	return &newMap
}

// GameModel methods --------------------------------------------------------------------------

// RegisterPlayer adds a new player to our world model
func (gm *GameModel) RegisterPlayer(ws *websocket.Conn) *Player {
	newP := &Player{
		Name: "",
		// Team:           nil,
		socket:   ws,
		outgoing: make(chan Message),
	}

	gm.Players[newP.ID] = newP
	return newP
}

func (gm *GameModel) broadcastState() {
	for _, player := range gm.Players {
		player.outgoing <- calcStateMsgForPlayer(player)
	}
}

func (gm *GameModel) setPlayerName(p *Player, n string) error {
	// check to see if name is in use
	for _, player := range gm.Players {
		if player.Name == playerName(n) {
			return errors.New("Name '" + n + "' already in use")
		}
	}

	// if not set it and return no error
	p.Name = playerName(n)
	return nil
}

func (gm *GameModel) setPlayerPOE(p *Player, n nodeID) bool {
	// TODO move this node validity check to a nodeMap method
	// if nodeID is valid

	if gm.Map.nodeExists(n) {
		gm.POEs[p.ID] = gm.Map.Nodes[n]

		//Just here for debugging TODO
		gm.Map.Nodes[n].addModule(newModuleBy(p))
		return true
	}

	return false
}

// RemovePlayer ...
func (gm *GameModel) RemovePlayer(p *Player) error {
	if _, ok := gm.Players[p.ID]; !ok {
		return errors.New("player '" + string(p.Name) + "' is not registered")
	}

	if p.Team != nil {
		p.Team.removePlayer(p)
	}

	// clean up POE
	delete(gm.POEs, p.ID)

	// Clean up route
	delete(gm.Routes, p.ID)

	// Clean up player
	delete(gm.Players, p.ID)

	return nil
}

func (gm *GameModel) assignPlayerToTeam(p *Player, tn teamName) error {
	if team, ok := gm.Teams[tn]; !ok {
		return errors.New("The team: " + string(tn) + " does not exist")
	} else if p.Team != nil {
		return errors.New(string(p.Name) + " is alread a member of team: " + string(tn))
	} else if team.isFull() {
		return errors.New("team: " + string(tn) + " is full")
	}

	gm.Teams[tn].addPlayer(p)
	return nil
}

func (gm *GameModel) tryConnectPlayerToNode(p *Player, n nodeID) bool {
	log.Printf("player %v attempting to connect to node %v from POE %v", p.Name, n, gm.POEs[p.ID])

	// if player is connected elsewhere, break that first, regardless of success of this attempt
	if gm.Routes[p.ID] != nil {
		gm.breakConnection(p)
	}

	// TODO handle player connecting to own POE

	source := gm.POEs[p.ID]
	target := gm.Map.Nodes[n]

	route := gm.Map.routeToNode(p, source, target)
	if route != nil {
		log.Println("Successful Connect")
		gm.establishConnection(p, route, target) // This should add player traffic to each intermediary and establish a connection on n
		return true
	}
	log.Println("Cannot Connect")
	return false
}

// TODO should this have gm as receiver? there's no need but makes sense syntactically
func (gm *GameModel) establishConnection(p *Player, routeNodes []*node, n *node) {
	// set's players route to the route generated via routeToNode
	gm.Routes[p.ID] = &route{Endpoint: n, Nodes: routeNodes}

}

func (gm *GameModel) breakConnection(p *Player) {
	if gm.Routes[p.ID] != nil {
		gm.Routes[p.ID] = nil
	}
}

// module methods -------------------------------------------------------------------------

func (m module) isFriendlyTo(p *Player) bool {
	if m.Builder.Team == p.Team {
		return true
	}
	return false
}

// node methods -------------------------------------------------------------------------------

// addConnection is reciprocol
func (n *node) addConnection(m *node) {
	n.Connections = append(n.Connections, m.ID)
	m.Connections = append(m.Connections, n.ID)
}

func (n *node) allowsRoutingFor(p *Player) bool {
	for _, module := range n.Modules {
		if module.isFriendlyTo(p) {
			return true
		}
	}
	return false
}

func (n *node) addModule(m module) {
	// TODO QUESTION do I need to return bool since failure is possible?
	if len(n.Modules) < n.Size {
		n.Modules[m.ID] = m
	}
}

func (n *node) removeModule(m module) {
	delete(n.Modules, m.ID)
}

// helper function for removing item from slice
// func cutPlayer(s []*Player, p *Player) []*Player {
// 	for i, thisP := range s {
// 		if p == thisP {
// 			// swaps the last element with the found element and returns with the last element cut
// 			s[len(s)-1], s[i] = s[i], s[len(s)-1]
// 			return s[:len(s)-1]
// 		}
// 	}
// 	log.Printf("CutPlayer returning: %v", s)
// 	return s
// }

// nodeMap methods -----------------------------------------------------------------------------

func (m *nodeMap) addNodes(ns ...*node) {
	for _, node := range ns {
		m.Nodes = append(m.Nodes, node)
	}
}

func (m *nodeMap) connectNodes(n1, n2 nodeID) error {
	// Check existence of both elements
	if m.nodeExists(n1) && m.nodeExists(n2) {

		// add connection value to each node,
		m.Nodes[n1].addConnection(m.Nodes[n2])
		return nil

	}

	log.Println("connectNodes error")
	return errors.New("One or both nodes out of range")
}

func (m *nodeMap) nodeExists(n nodeID) bool {
	if n > -1 && n < len(m.Nodes) {
		return true
	}
	return false
}

// nodesConnections takes one of the maps nodes and converts its connections (in the form of nodeIDs) into pointers to actual node objects
// TODO ask about this, feels hacky
func (m *nodeMap) nodesConnections(n *node) []*node {
	res := make([]*node, 0)
	for _, nodeID := range n.Connections {
		res = append(res, m.Nodes[nodeID])
	}

	return res
}

func (m *nodeMap) nodesTouch(n1, n2 *node) bool {
	// for every one of n1's connections
	for _, connectedNode := range m.nodesConnections(n1) {
		// if it is n2, return true
		if connectedNode == n2 {
			return true
		}
	}
	return false
}

// routeToNode uses vanilla dijkstra's (vanilla for now) algorithm to find node path
func (m *nodeMap) routeToNode(p *Player, source, target *node) []*node {

	// if we're connecting to our POE, return a route which is only our POE
	if source == target {
		if source.allowsRoutingFor(p) {
			route := make([]*node, 1)
			route[0] = source
			return route
		}
		log.Println("POE Blocked")
		return nil

	}
	unchecked := make(map[*node]bool) // TODO this should be a priority queue for efficiency
	dist := make(map[*node]int)
	prev := make(map[*node]*node)

	for _, node := range m.Nodes {
		// Only consider node if node is friendly to player (i.e. has module from team)
		if node.allowsRoutingFor(p) {
			dist[node] = 10000
			unchecked[node] = true
		}
	}

	dist[source] = 0

	for len(unchecked) > 0 {
		thisNode := getBestNode(unchecked, dist)

		delete(unchecked, thisNode)

		if m.nodesTouch(thisNode, target) {
			route := constructPath(prev, target)
			log.Println("Found target!")
			log.Printf("%v", route)
			return route
		}

		for _, cNode := range m.nodesConnections(thisNode) {
			alt := dist[thisNode] + 1
			if alt < dist[cNode] {
				dist[cNode] = alt
				prev[cNode] = thisNode
			}
		}
	}
	log.Println("No possible route")
	return nil
}

// helper functions for routeToNode ------------------------------------------------------------
// constructPath takes the routes discovered via routeToNode and the endpoint (target) and creates a slice of the correct path, note order is still reversed and path contains source but not target node
func constructPath(prevMap map[*node]*node, t *node) []*node {
	// log.Println(prevMap)

	route := make([]*node, 0)

	for step, ok := prevMap[t]; ok; step, ok = prevMap[step] {
		route = append(route, step)
	}

	return route
}

// getBestNode TODO extract the node with shortes path from pool, it is a substitute for using a priority queue
func getBestNode(pool map[*node]bool, distMap map[*node]int) *node {
	bestDist := 100000
	var bestNode *node
	for node := range pool {
		if distMap[node] < bestDist {
			bestNode = node
			bestDist = distMap[node]
		}
	}
	return bestNode
}

// player methods -------------------------------------------------------------------------------

func newPlayer(ws *websocket.Conn) *Player {
	ret := &Player{
		ID:   playerIDCount,
		Name: "",
		// Team:           nil,
		socket:   ws,
		outgoing: make(chan Message),
	}
	playerIDCount++
	return ret
}

func (p Player) hasTeam() bool {
	if p.Team == nil {
		return false
	}
	return true
}

func (p Player) hasName() bool {
	if p.Name == "" {
		return false
	}
	return true
}

// route methods --------------------------------------------
// func (r route) isActive() bool {
// 	if r.Endpoint == -1 {
// 		return false
// 	}
// 	return true
// }

// func (r *route) terminate() {
// 	r.Endpoint = -1
// }

// team methods -------------------------------------------------------------------------------
func (t team) isFull() bool {
	if len(t.players) < t.MaxSize {
		return false
	}
	return true
}

func (t *team) broadcast(msg Message) {
	for player := range t.players {
		player.outgoing <- msg
	}
}

func (t *team) addPlayer(p *Player) {
	t.players[p] = true
	p.Team = t

	// TODO Tell client they've joined model shouldn't handle messaging, fix
	p.outgoing <- Message{
		Type:   "teamAssign",
		Sender: "server",
		Data:   string(t.Name),
	}
}

func (t *team) removePlayer(p *Player) {
	delete(t.players, p)
	p.Team = nil

	// Notify client model shouldn't handle messaging, fix
	p.outgoing <- Message{
		Type:   "teamUnassign",
		Sender: "server",
		Data:   string(t.Name),
	}
}

// Stringers ----------------------------------------------------------------------------------

func (n node) String() string {
	return fmt.Sprintf("<(node) ID: %v, Connections:%v, Modules:%v>", n.ID, n.Connections, n.Modules)
}

func (t team) String() string {
	var playerList []string
	for player := range t.players {
		playerList = append(playerList, string(player.Name))
	}
	return fmt.Sprintf("<team> (Name: %v, Players:%v)", t.Name, playerList)
}

func (p Player) String() string {
	return fmt.Sprintf("<player> Name: %v, team: %v", p.Name, p.Team)
}