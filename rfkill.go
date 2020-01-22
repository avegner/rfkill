//nolint:wsl,gomnd,gocognit
package rfkill

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func List(ctx context.Context) ([]*Device, error) {
	fd, err := openEventDev()
	if err != nil {
		return nil, err
	}
	defer syscall.Close(fd)

	devs := []*Device{}
	var raw [rfkillEventSize]byte

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if _, err = syscall.Read(fd, raw[:]); err != nil {
			if err != syscall.EAGAIN {
				return nil, err
			}
			return devs, nil
		}

		ev := rfkillEvent{}
		ev.unmarshall(raw)

		switch op := EventOp(ev.Op); op {
		case AddOp:
			name, err := getDevName(ev.ID)
			if err != nil {
				return nil, err
			}
			dev := &Device{
				ID:   ev.ID,
				Type: RadioType(ev.Type),
				Name: name,
			}
			updateDevState(dev, &ev)
			devs = append(devs, dev)
		case DelOp:
			for i, dev := range devs {
				if dev.ID == ev.ID {
					devs = append(devs[0:i], devs[i+1:]...)
				}
			}
		case ChangeOp:
			for _, dev := range devs {
				if dev.ID == ev.ID {
					updateDevState(dev, &ev)
				}
			}
		case ChangeAllOp:
			for _, dev := range devs {
				if RadioType(ev.Type) == AllRadio || RadioType(ev.Type) == dev.Type {
					updateDevState(dev, &ev)
				}
			}
		default:
			return nil, fmt.Errorf("unknown rfkill event %d", ev.Op)
		}
	}
}

func Block(option BlockOption) error {
	return block(1, option)
}

func Unblock(option BlockOption) error {
	return block(0, option)
}

func Events(ctx context.Context, pollInterval time.Duration, callback func(ev *Event)) error {
	fd, err := openEventDev()
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	var raw [rfkillEventSize]byte

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err = syscall.Read(fd, raw[:]); err != nil {
			if err != syscall.EAGAIN {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
				continue
			}
		}

		ev := rfkillEvent{}
		ev.unmarshall(raw)

		callback(&Event{
			ID:        ev.ID,
			Op:        EventOp(ev.Op),
			Type:      RadioType(ev.Type),
			HardBlock: ev.Hard > 0,
			SoftBlock: ev.Soft > 0,
		})
	}
}

func openEventDev() (int, error) {
	// device doesn't return EOF
	return syscall.Open("/dev/rfkill", syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
}

func getDevName(id uint32) (string, error) {
	name, err := ioutil.ReadFile("/sys/class/rfkill/rfkill" +
		strconv.FormatUint(uint64(id), 10) + "/name")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(name), "\n"), nil
}

func block(soft uint8, option BlockOption) error {
	ev := rfkillEvent{
		Soft: soft,
	}
	option(&ev)

	f, err := os.OpenFile("/dev/rfkill", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	m := ev.marshall()
	_, err = f.Write(m[:])
	return err
}

func updateDevState(dev *Device, ev *rfkillEvent) {
	dev.HardBlock = ev.Hard > 0
	dev.SoftBlock = ev.Soft > 0
}

type Device struct {
	ID        uint32
	Type      RadioType
	HardBlock bool
	SoftBlock bool
	Name      string
}

type Event struct {
	ID        uint32
	Op        EventOp
	Type      RadioType
	HardBlock bool
	SoftBlock bool
}

type RadioType uint8

const (
	AllRadio RadioType = iota
	WLANRadio
	BluetoothRadio
	UWBRadio
	WIMAXRadio
	WWANRadio
	GPSRadio
	FMRadio
	NFCRadio
)

type EventOp uint8

const (
	AddOp EventOp = iota
	DelOp
	ChangeOp
	ChangeAllOp
)

type BlockOption func(ev *rfkillEvent)

func WithID(id uint) BlockOption {
	return func(ev *rfkillEvent) {
		ev.Op = uint8(ChangeOp)
		ev.ID = uint32(id)
	}
}

func WithType(typ RadioType) BlockOption {
	return func(ev *rfkillEvent) {
		ev.Op = uint8(ChangeAllOp)
		ev.Type = uint8(typ)
	}
}

type rfkillEvent struct {
	ID   uint32
	Type uint8
	Op   uint8
	Soft uint8
	Hard uint8
}

const rfkillEventSize = 8

func (ev *rfkillEvent) marshall() [rfkillEventSize]byte {
	bs := [8]byte{
		0x00, 0x00, 0x00, 0x00,
		ev.Type,
		ev.Op,
		ev.Soft,
		ev.Hard,
	}
	endianess.PutUint32(bs[:], ev.ID)
	return bs
}

func (ev *rfkillEvent) unmarshall(bs [rfkillEventSize]byte) {
	ev.ID = endianess.Uint32(bs[:])
	ev.Type = bs[4]
	ev.Op = bs[5]
	ev.Soft = bs[6]
	ev.Hard = bs[7]
}
