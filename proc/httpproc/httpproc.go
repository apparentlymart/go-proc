package httpproc

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/apparentlymart/go-proc/proc"
)

// Serve returns a process implementation that serves the given HTTP server
// on the given listener until the process context is cancelled or hits its
// deadline.
//
// Once the context is done, it will instruct the server to shut down and
// return only after shutdown has completed.
//
// The Handler field in srv is replaced with a wrapper that attaches the
// process context to the request.
//
// The process will fail (return an error) if listening fails or if shutdown
// fails. A successful shutdown is not an error.
func Serve(srv *http.Server, l net.Listener) proc.Impl {
	return server(srv, func() error {
		return srv.Serve(l)
	})
}

// ServeTLS is like Serve but runs a TLS server instead. The TLSConfig field
// of srv and the certFile and keyFile arguments are used in the same way
// as for http.ServeTLS.
func ServeTLS(srv *http.Server, l net.Listener, certFile, keyFile string) proc.Impl {
	return server(srv, func() error {
		return srv.ServeTLS(l, certFile, keyFile)
	})
}

// ListenAndServe is like Serve but it first creates a listen socket for
// srv.Addr and serves on that. Along with the error conditions for Serve,
// ListenAndServe may also return an error if listening fails.
func ListenAndServe(srv *http.Server) proc.Impl {
	return server(srv, func() error {
		return srv.ListenAndServe()
	})
}

// ListenAndServeTLS is like ServeTLS but it first creates a listen socket for
// srv.Addr and serves on that. Along with the error conditions for Serve,
// ListenAndServe may also return an error if listening fails. certFile and
// keyFile have the same meaning as for http.ListenAndServeTLS.
func ListenAndServeTLS(srv *http.Server, certFile, keyFile string) proc.Impl {
	return server(srv, func() error {
		return srv.ListenAndServeTLS(certFile, keyFile)
	})
}

func server(srv *http.Server, serve func() error) proc.Impl {
	return func(ctx context.Context) error {
		realHandler := srv.Handler
		srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req = req.WithContext(ctx)
			realHandler.ServeHTTP(w, req)
		})

		var serveErr error
		var wg sync.WaitGroup
		wg.Add(1) // the server itself
		go func() {
			serveErr = serve()
			wg.Done()
		}()

		<-ctx.Done()
		// Since our context is already cancelled, we'll use a background one
		// for shutdown. Since there's no deadline here it will wait indefinitely
		// for requests to exit, under the assumption that they will be written
		// to watch for context cancellation themselves and abort any
		// long-running work.
		shutdownErr := srv.Shutdown(context.Background())
		wg.Wait() // wait for the server to have exited, though in the normal path it will exit before Shutdown returns

		if serveErr != nil {
			return serveErr
		}
		return shutdownErr
	}
}
