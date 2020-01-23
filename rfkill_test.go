//nolint:wsl,gochecknoinits,gochecknoglobals,gomnd
package rfkill_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/avegner/rfkill"
)

var (
	idToBlock   uint = 1
	typeToBlock      = rfkill.AllRadio
)

func init() {
	getEnv()
}

func TestList(t *testing.T) {
	devs, err := rfkill.List(context.Background())
	if err != nil {
		t.Fatalf("got err '%v', want nil err", err)
	}
	if devs == nil {
		t.Fatal("got nil devs, want allocated slice")
	}
}

func TestBlockWithID(t *testing.T) {
	if err := rfkill.Block(rfkill.WithID(idToBlock)); err != nil {
		t.Fatalf("got err '%v', want nil err", err)
	}
}

func TestUnblockWithID(t *testing.T) {
	if err := rfkill.Unblock(rfkill.WithID(idToBlock)); err != nil {
		t.Fatalf("got err '%v', want nil err", err)
	}
}

func TestBlockWithType(t *testing.T) {
	if err := rfkill.Block(rfkill.WithType(typeToBlock)); err != nil {
		t.Fatalf("got err '%v', want nil err", err)
	}
}

func TestUnblockWithType(t *testing.T) {
	if err := rfkill.Unblock(rfkill.WithType(typeToBlock)); err != nil {
		t.Fatalf("got err '%v', want nil err", err)
	}
}

func TestEvents(t *testing.T) {
	errc := make(chan error, 1)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
	defer cancel()

	go func() {
		errc <- rfkill.Events(ctx, 100*time.Millisecond, func(ev *rfkill.Event) {
			t.Logf("event: %+v", *ev)
		})
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("func hasn't returned after deadline context exceeded")
	case err := <-errc:
		if err != context.DeadlineExceeded {
			t.Fatalf("got err '%v', want err '%v'", err, context.DeadlineExceeded)
		}
	}
}

func BenchmarkList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := rfkill.List(context.Background()); err != nil {
			b.Fatalf("list err: %v", err)
		}
	}
}

func BenchmarkBlockWithID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := rfkill.Block(rfkill.WithID(idToBlock)); err != nil {
			b.Fatalf("block err: %v", err)
		}
		if err := rfkill.Unblock(rfkill.WithID(idToBlock)); err != nil {
			b.Fatalf("unblock err: %v", err)
		}
	}
}

func BenchmarkBlockWithType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := rfkill.Block(rfkill.WithType(typeToBlock)); err != nil {
			b.Fatalf("block err: %v", err)
		}
		if err := rfkill.Unblock(rfkill.WithType(typeToBlock)); err != nil {
			b.Fatalf("unblock err: %v", err)
		}
	}
}

func getEnv() {
	if s := os.Getenv("ID"); s != "" {
		id, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			panic(fmt.Sprintf("ID parse: %v", err))
		}
		idToBlock = uint(id)
	}
	if s := os.Getenv("TYPE"); s != "" {
		typ, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			panic(fmt.Sprintf("TYPE parse: %v", err))
		}
		typeToBlock = rfkill.RadioType(typ)
	}
}
