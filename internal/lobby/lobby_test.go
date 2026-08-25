package lobby

import "testing"

func pad(id int) Device {
	return Device{Kind: DeviceGamepad, ID: id, Name: "Test pad"}
}

func TestJoinAssignsUpToFourUniqueDevices(t *testing.T) {
	var l Lobby
	for id := 0; id < MaxPlayers; id++ {
		slot, joined := l.Join(pad(id))
		if !joined || slot != id {
			t.Fatalf("join %d = (%d, %v)", id, slot, joined)
		}
	}
	if slot, joined := l.Join(pad(4)); joined || slot != -1 {
		t.Fatalf("fifth join = (%d, %v)", slot, joined)
	}
}

func TestJoinIsIdempotent(t *testing.T) {
	var l Lobby
	l.Join(pad(7))
	if slot, joined := l.Join(pad(7)); joined || slot != 0 {
		t.Fatalf("duplicate join = (%d, %v)", slot, joined)
	}
	if len(l.Slots) != 1 {
		t.Fatalf("got %d slots", len(l.Slots))
	}
}

func TestCanStartWithTwoPlayers(t *testing.T) {
	var l Lobby
	l.Join(pad(1))
	if l.CanStart() {
		t.Fatal("one player should not start")
	}
	l.Join(pad(2))
	if !l.CanStart() {
		t.Fatal("two players should start")
	}
}

func TestLeaveRemovesAssignedDevice(t *testing.T) {
	var l Lobby
	l.Join(pad(1))
	l.Join(pad(2))
	if !l.Leave(pad(1)) || len(l.Slots) != 1 || l.Slots[0].Device.ID != 2 {
		t.Fatalf("unexpected slots after leave: %+v", l.Slots)
	}
}
