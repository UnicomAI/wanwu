package mcpoauth

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestFileStoreProcessWorker(t *testing.T) {
	dir := os.Getenv("WANWU_TEST_OAUTH_STORE")
	if dir == "" {
		t.Skip("subprocess helper")
	}
	s := diskStore(t, dir)
	for i := 0; i < 10; i++ {
		if e := s.Update(context.Background(), Owner{"org", "alice"}, "https://resource.example/mcp", func(g *Grant) (*Grant, error) {
			if g == nil {
				g = &Grant{}
			}
			n, _ := strconv.Atoi(g.Generation)
			g.Generation = strconv.Itoa(n + 1)
			return g, nil
		}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestFileStoreSerializesProcesses(t *testing.T) {
	dir := t.TempDir()
	s := diskStore(t, dir)
	commands := []*exec.Cmd{}
	for i := 0; i < 4; i++ {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFileStoreProcessWorker$")
		cmd.Env = append(os.Environ(), "WANWU_TEST_OAUTH_STORE="+dir)
		if e := cmd.Start(); e != nil {
			t.Fatal(e)
		}
		commands = append(commands, cmd)
	}
	for _, cmd := range commands {
		if e := cmd.Wait(); e != nil {
			t.Fatal(e)
		}
	}
	if e := s.Update(context.Background(), Owner{"org", "alice"}, "https://resource.example/mcp", func(g *Grant) (*Grant, error) {
		if g == nil || g.Generation != "40" {
			t.Fatalf("lost cross-process updates: %+v", g)
		}
		return g, nil
	}); e != nil {
		t.Fatal(e)
	}
}
