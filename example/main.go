package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"rfkill"
)

var errUsage = errors.New("invalid usage")

var types = map[string]rfkill.RadioType{
	"all":       rfkill.AllRadio,
	"wlan":      rfkill.WLANRadio,
	"bluetooth": rfkill.BluetoothRadio,
	"uwb":       rfkill.UWBRadio,
	"wimax":     rfkill.WIMAXRadio,
	"wwan":      rfkill.WWANRadio,
	"gps":       rfkill.GPSRadio,
	"fm":        rfkill.FMRadio,
	"nfc":       rfkill.NFCRadio,
}

var mlog = log.New(os.Stderr, "", 0)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage:

	%s <command> [argument...]

The commands are:

	list
	block-id      <id>    block radio with ID
	unblock-id    <id>    unblock radio with ID
	block-type    <type>  block all radios with type
	unblock-type  <type>  unblock all radios with type
	events                track events

The types are:

	all        all radios
	wlan       WLAN radio
	bluetooth  Bluetooth radio
	uwb        UWB radio
	wimax      WIMAX radio
	wwan       WWAN radio
	gps        GPS radio
	fm         FM radio
	nfc        NFC radio

`, path.Base(os.Args[0]))
	}
	flag.Parse()

	if err := run(); err != nil {
		if err == errUsage {
			flag.Usage()
			os.Exit(2)
		}
		mlog.Fatalf("err: %v", err)
	}
}

//nolint:gocognit
func run() error {
	if flag.NArg() < 1 {
		return errUsage
	}

	switch cmd := flag.Arg(0); cmd {
	case "list":
		devs, err := rfkill.List()
		if err != nil {
			return err
		}
		for i, dev := range devs {
			mlog.Printf("%d: %+v", i, *dev)
		}
	case "block-id", "unblock-id":
		if flag.NArg() < 2 {
			return errUsage
		}
		id, err := strconv.ParseUint(os.Args[2], 10, 32)
		if err != nil {
			return err
		}
		if cmd == "block-id" {
			if err := rfkill.Block(rfkill.WithID(uint(id))); err != nil {
				return err
			}
		} else {
			if err := rfkill.Unblock(rfkill.WithID(uint(id))); err != nil {
				return err
			}
		}
	case "block-type", "unblock-type":
		if flag.NArg() < 2 {
			return errUsage
		}
		typ, ok := types[flag.Arg(1)]
		if !ok {
			return errUsage
		}
		if cmd == "block-type" {
			if err := rfkill.Block(rfkill.WithType(typ)); err != nil {
				return err
			}
		} else {
			if err := rfkill.Unblock(rfkill.WithType(typ)); err != nil {
				return err
			}
		}
	case "events":
		// blocks forever
		if err := rfkill.Events(context.Background(), 1*time.Second, func(ev *rfkill.Event) {
			mlog.Printf("event: %+v", *ev)
		}); err != nil {
			return err
		}
	default:
		return errUsage
	}
	return nil
}
