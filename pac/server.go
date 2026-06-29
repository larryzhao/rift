package pac

import (
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	Port        int
	DomainsFile string
	GFWListFile string
}

func NewServer(port int, domainsFile, gfwlistFile string) *Server {
	return &Server{Port: port, DomainsFile: domainsFile, GFWListFile: gfwlistFile}
}

func (server *Server) Run() error {
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", server.Port),
	}

	http.HandleFunc("/pac/proxy.js", func(w http.ResponseWriter, r *http.Request) {
		js, err := GeneratePAC(server.DomainsFile, server.GFWListFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("generate pac err: " + err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(js))
	})

	defer s.Shutdown(context.Background())
	return s.ListenAndServe()
}
