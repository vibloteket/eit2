// Package lobby assigns input devices to local player slots.
package lobby

import "fmt"

const MaxPlayers = 4

type DeviceKind string

const (
	DeviceKeyboard DeviceKind = "keyboard"
	DeviceGamepad  DeviceKind = "gamepad"
	DeviceTouch    DeviceKind = "touch"
)

type Device struct {
	Kind DeviceKind
	ID   int
	Name string
}

func (d Device) Key() string {
	return fmt.Sprintf("%s:%d", d.Kind, d.ID)
}

type Slot struct {
	Device Device
	Ready  bool
}

type Lobby struct {
	Slots []Slot
}

func (l *Lobby) Join(device Device) (int, bool) {
	for i, slot := range l.Slots {
		if slot.Device.Key() == device.Key() {
			return i, false
		}
	}
	if len(l.Slots) >= MaxPlayers {
		return -1, false
	}
	l.Slots = append(l.Slots, Slot{Device: device, Ready: true})
	return len(l.Slots) - 1, true
}

func (l *Lobby) PlayerForDevice(device Device) int {
	for i, slot := range l.Slots {
		if slot.Device.Key() == device.Key() {
			return i
		}
	}
	return -1
}

func (l *Lobby) ReplaceDevice(player int, device Device) bool {
	if player < 0 || player >= len(l.Slots) {
		return false
	}
	existing := l.PlayerForDevice(device)
	if existing >= 0 && existing != player {
		return false
	}
	l.Slots[player].Device = device
	l.Slots[player].Ready = true
	return true
}

func (l *Lobby) Leave(device Device) bool {
	for i, slot := range l.Slots {
		if slot.Device.Key() == device.Key() {
			l.Slots = append(l.Slots[:i], l.Slots[i+1:]...)
			return true
		}
	}
	return false
}

func (l *Lobby) CanStart() bool {
	return len(l.Slots) >= 1
}
