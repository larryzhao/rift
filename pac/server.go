package pac

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

type Server struct {
	Port    int
	PACFile string
}

func NewServer(port int, pacFile string) *Server {
	return &Server{Port: port, PACFile: pacFile}
}

func (server *Server) Run() error {
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", server.Port),
	}

	http.HandleFunc("/pac/proxy.js", func(w http.ResponseWriter, r *http.Request) {
		bb, err := os.ReadFile(server.PACFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("read pac file err: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bb)
	})

	defer s.Shutdown(context.Background())
	return s.ListenAndServe()
}
