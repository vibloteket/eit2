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

func New(playerCount int) *Match {
	m := &Match{
		Players: make([]*core.Game, playerCount),
		Targets: make([]int, playerCount),
		Winner:  -1,
	}
	for i := range m.Players {
		m.Players[i] = core.New(uint64(i + 1))
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
	m.UpdateStatus()
}

func (m *Match) routeGarbage() {
	for attacker, player := range m.Players {
		rows := player.ConsumeGarbage()
		if rows == 0 {
			continue
		}
		target := m.Target(attacker)
		if target >= 0 && target < len(m.Players) && !m.Players[target].GameOver {
			m.Players[target].AddGarbage(rows)
		}
	}
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
		target := m.Targets[i]
		if target < 0 || target >= len(m.Players) || m.Players[target].GameOver || target == i {
			m.Targets[i] = m.nextAlive(i, i)
		}
	}
}

func (m *Match) nextAlive(player, after int) int {
	if len(m.Players) < 2 {
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
