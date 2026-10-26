package probe

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// fakeTLD serves "taken.test." as delegated, "slow.test." after a delay
// longer than the client timeout, and NXDOMAIN for everything else. It
// records the client addresses it saw.
type fakeTLD struct {
	mu    sync.Mutex
	ports map[string]int
}

func (f *fakeTLD) ServeDNS(w dns.ResponseWriter, q *dns.Msg) {
	f.mu.Lock()
	f.ports[w.RemoteAddr().String()]++
	f.mu.Unlock()
	name := q.Question[0].Name
	if strings.HasPrefix(name, "slow.") {
		time.Sleep(300 * time.Millisecond)
	}
	r := new(dns.Msg)
	r.SetReply(q)
	if !strings.HasPrefix(name, "taken.") && !strings.HasPrefix(name, "slow.") {
		r.Rcode = dns.RcodeNameError
	}
	w.WriteMsg(r)
}

func startFake(t *testing.T) (*fakeTLD, string) {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeTLD{ports: map[string]int{}}
	srv := &dns.Server{PacketConn: pc, Handler: f}
	go srv.ActivateAndServe()
	t.Cleanup(func() { srv.Shutdown() })
	return f, pc.LocalAddr().String()
}

func TestCheckReusesSockets(t *testing.T) {
	f, addr := startFake(t)
	p := newProber("test", []string{addr}, 10000)
	defer p.Close()
	ctx := context.Background()

	if r := p.Check(ctx, "taken"); r != Taken {
		t.Fatalf("taken: %s", r)
	}
	const n = 200
	for i := 0; i < n; i++ {
		if r := p.Check(ctx, "free"); r != Free {
			t.Fatalf("free #%d: %s", i, r)
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ports) > 2 {
		t.Errorf("%d sequential queries used %d sockets, want reuse", n+1, len(f.ports))
	}
}

func TestCheckConcurrent(t *testing.T) {
	f, addr := startFake(t)
	p := newProber("test", []string{addr}, 10000)
	defer p.Close()
	var wg sync.WaitGroup
	for w := 0; w < 16; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				label, want := "free", Free
				if i%3 == 0 {
					label, want = "taken", Taken
				}
				if r := p.Check(context.Background(), label); r != want {
					t.Errorf("%s: %s, want %s", label, r, want)
					return
				}
			}
		}()
	}
	wg.Wait()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ports) > 16+idlePerServer {
		t.Errorf("800 queries from 16 workers used %d sockets", len(f.ports))
	}
}

// BenchmarkCheck compares pooled sockets with one socket per query (the
// previous behaviour) against a local server.
func BenchmarkCheck(b *testing.B) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	srv := &dns.Server{PacketConn: pc, Handler: &fakeTLD{ports: map[string]int{}}}
	go srv.ActivateAndServe()
	defer srv.Shutdown()
	addr := pc.LocalAddr().String()
	p := newProber("test", []string{addr}, 1e9)
	defer p.Close()
	m := new(dns.Msg)
	m.SetQuestion("free.test.", dns.TypeNS)
	b.Run("pooled", func(b *testing.B) {
		b.SetParallelism(4)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				p.exchange(context.Background(), m, addr)
			}
		})
	})
	b.Run("socket-per-query", func(b *testing.B) {
		b.SetParallelism(4)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				p.udp.ExchangeContext(context.Background(), m, addr)
			}
		})
	})
}

// A timed-out query must not leave its socket in the pool, where the late
// reply would be read by the next query.
func TestTimeoutDiscardsSocket(t *testing.T) {
	_, addr := startFake(t)
	p := newProber("test", []string{addr}, 10000)
	p.udp.Timeout = 100 * time.Millisecond
	p.retries = 1
	defer p.Close()
	if r := p.Check(context.Background(), "slow"); r != Unknown {
		t.Fatalf("slow: %s, want unknown", r)
	}
	if n := len(p.idle[addr]); n != 0 {
		t.Fatalf("%d sockets pooled after a timeout", n)
	}
	time.Sleep(400 * time.Millisecond) // the late reply arrives now
	if r := p.Check(context.Background(), "free"); r != Free {
		t.Fatalf("free after timeout: %s", r)
	}
}
