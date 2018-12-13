package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/trace"
	"sync/atomic"

	"github.com/apparentlymart/go-proc/proc"
	"github.com/apparentlymart/go-proc/proc/httpproc"
)

func main() {
	// Enable tracing
	traceF, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to open trace file: %s", err)
	}
	err = trace.Start(traceF)
	if err != nil {
		log.Fatalf("failed to start tracing: %s", err)
	}
	defer trace.Stop()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: &HitCounter{},
	}

	// Label the web server process as a task, so its trace events will
	// be annotated.
	web := proc.Task("web", httpproc.ListenAndServe(srv))

	// Running the server as a process in proc.Main means Ctrl+C will
	// request a graceful shutdown of the server.
	err = proc.Main(web)

	if err != nil {
		log.Fatal(err)
	}
}

type HitCounter struct {
	count int64
}

func (h *HitCounter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, task := trace.NewTask(req.Context(), "request")
	defer task.End()

	if req.URL.Path != "/" {
		w.WriteHeader(404)
		fmt.Fprint(w, "Not found")
		trace.Log(ctx, "responseStatus", "404")
		return
	}

	new := atomic.AddInt64(&h.count, 1)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	trace.Log(ctx, "responseStatus", "200")
	trace.Logf(ctx, "newCount", "%d", new)
	switch new {
	case 1:
		fmt.Fprint(w, "<p>This server has been accessed only once.</p>")
	case 2:
		fmt.Fprint(w, "<p>This server has been accessed twice.</p>")
	default:
		fmt.Fprintf(w, "<p>This server has been accessed %d times.</p>", new)
	}
}
