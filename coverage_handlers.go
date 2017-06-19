package main

import (
	"github.com/archivers-space/api/apiutil"
	"github.com/archivers-space/archive"
	"github.com/archivers-space/coverage/coverage"
	"github.com/archivers-space/coverage/tree"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

func CoverageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		EmptyOkHandler(w, r)
	case "GET":
		CoverageTreeHandler(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func CoverageTreeHandler(w http.ResponseWriter, r *http.Request) {
	var primer *archive.Primer
	patterns := strings.Split(r.FormValue("patterns"), ",")
	if r.FormValue("primer") != "" {
		primer = &archive.Primer{Id: r.FormValue("primer")}
		if err := primer.Read(store); err != nil {
			apiutil.WriteErrResponse(w, http.StatusBadRequest, err)
			return
		}
		if err := primer.ReadSources(appDB); err != nil {
			apiutil.WriteErrResponse(w, http.StatusBadRequest, err)
			return
		}
		patterns = make([]string, len(primer.Sources))
		for i, s := range primer.Sources {
			patterns[i] = s.Url
		}
	}

	conn, err := net.Dial("tcp", cfg.CoverageServiceUrl)
	if err != nil {
		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	cli := rpc.NewClient(conn)

	depth, err := apiutil.ReqParamInt("depth", r)
	if err != nil {
		depth = 0
	}

	p := coverage.CoverageTreeParams{
		Patterns: patterns,
		Depth:    depth,
	}
	reply := &tree.Node{}
	if err := cli.Call("CoverageRequests.Tree", p, reply); err != nil {
		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	if primer != nil {
		reply.Name = primer.Title
	}
	apiutil.WriteResponse(w, reply)
}

// func CoverageSummaryHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := net.Dial("tcp", cfg.CoverageServiceUrl)
// 	if err != nil {
// 		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	cli := rpc.NewClient(conn)
// 	p := coverage.CoverageSummaryParams{
// 		Patterns: strings.Split(r.FormValue("patterns"), ","),
// 	}
// 	reply := &coverage.Summary{}
// 	if err := cli.Call("CoverageRequests.Summary", p, reply); err != nil {
// 		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
// 		return
// 	}
// 	apiutil.WriteResponse(w, reply)
// }
