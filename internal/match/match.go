// Package match coordinates multiple independent game boards.
package match

import core "github.com/vibloteket/eit2/internal/game"

// Match owns the player boards and multiplayer-only state.
type Match struct {
	Players []*core.Game
	Targets []int
	Over    bool
	Winner  int // -1 while there is no winner
}

// New creates a deterministic match for tests and recorded scenarios.
func New(playerCount int) *Match {
	return NewSeeded(playerCount, 1)
}

// NewSeeded creates a reproducible match while giving every player a distinct
// random stream derived from the match seed.
func NewSeeded(playerCount int, seed uint64) *Match {
	m := &Match{
		Players: make([]*core.Game, playerCount),
		Targets: make([]int, playerCount),
		Winner:  -1,
	}
	for i := range m.Players {
		playerSeed := seed + uint64(i)*0x9e3779b97f4a7c15
		m.Players[i] = core.New(playerSeed)
	}
	for i := range m.Targets {
		m.Targets[i] = m.nextAlive(i, i)
	}
	return m
}

func (m *Match) Tick() {
	if m.Over {
		return
	}
	for _, player := range m.Players {
		player.Tick()
	}
	m.routeGarbage()
	m.routeSpecials()
	m.UpdateStatus()
}

func (m *Match) routeGarbage() {
	for attacker, player := range m.Players {
		rows := player.ConsumeGarbage()
		if rows == 0 || len(m.Players) == 1 {
			continue
		}
		target := m.Target(attacker)
		if target >= 0 && target < len(m.Players) && !m.Players[target].GameOver {
			m.Players[target].AddGarbage(rows)
		}
	}
}

func (m *Match) routeSpecials() {
	for attacker, player := range m.Players {
		for _, special := range player.ConsumeSpecials() {
			target := m.Target(attacker)
			if len(m.Players) == 1 {
				target = attacker
			}
			if target < 0 || target >= len(m.Players) || m.Players[target].GameOver {
				continue
			}
			if special == core.SpecialSwitch {
				player.SwapBoard(m.Players[target])
				player.ShowEvent("SWITCHED WITH P" + playerNumber(target))
				if target != attacker {
					m.Players[target].ShowEvent("SWITCHED BY P" + playerNumber(attacker))
				}
				continue
			}
			m.Players[target].ApplySpecial(special)
			if target == attacker {
				player.ShowEvent("SELF: " + special.Name())
			} else {
				player.ShowEvent("SENT " + special.Name() + " TO P" + playerNumber(target))
				m.Players[target].ShowEvent("P" + playerNumber(attacker) + ": " + special.Name())
			}
		}
	}
}

func playerNumber(index int) string {
	return string(rune('1' + index))
}

// DebugCollect activates a special for a source player and immediately routes
// any target effect through normal match rules.
func (m *Match) DebugCollect(player int, special core.Special) bool {
	if player < 0 || player >= len(m.Players) || special == core.SpecialNone {
		return false
	}
	m.Players[player].CollectSpecial(special)
	m.routeSpecials()
	return true
}

func (m *Match) CycleTarget(player int) int {
	if player < 0 || player >= len(m.Players) || m.Players[player].GameOver {
		return -1
	}
	start := m.Targets[player]
	if start < 0 {
		start = player
	}
	m.Targets[player] = m.nextAlive(player, start)
	return m.Targets[player]
}

func (m *Match) Target(player int) int {
	if player < 0 || player >= len(m.Targets) {
		return -1
	}
	if len(m.Players) == 1 && !m.Players[player].GameOver {
		return player
	}
	return m.Targets[player]
}

func (m *Match) UpdateStatus() {
	alive := make([]int, 0, len(m.Players))
	for i, player := range m.Players {
		if !player.GameOver {
			alive = append(alive, i)
		}
	}
	if len(m.Players) > 1 && len(alive) <= 1 {
		m.Over = true
		if len(alive) == 1 {
			m.Winner = alive[0]
		}
	}
	for i, player := range m.Players {
		if player.GameOver {
			m.Targets[i] = -1
			continue
		}
		if len(m.Players) == 1 {
			m.Targets[i] = i
			continue
		}
		target := m.Targets[i]
		if target < 0 || target >= len(m.Players) || m.Players[target].GameOver || target == i {
			m.Targets[i] = m.nextAlive(i, i)
		}
	}
}

func (m *Match) nextAlive(player, after int) int {
	if len(m.Players) < 2 {
		if len(m.Players) == 1 && player == 0 && !m.Players[0].GameOver {
			return 0
		}
		return -1
	}
	for offset := 1; offset <= len(m.Players); offset++ {
		candidate := (after + offset) % len(m.Players)
		if candidate != player && !m.Players[candidate].GameOver {
			return candidate
		}
	}
	return -1
}
