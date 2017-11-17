package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/matryer/way"
)

// Server is the app server.
type Server struct {
	assets  string
	videos  string
	items   *Items // here the video items on the filesystem
	facebox *facebox.Client
	router  *way.Router
}

// NewServer makes a new Server.
func NewServer(assets string, videos string, facebox *facebox.Client) *Server {
	srv := &Server{
		assets:  assets,
		videos:  videos,
		items:   LoadItemsFromPath(videos),
		facebox: facebox,
		router:  way.NewRouter(),
	}

	srv.router.Handle(http.MethodGet, "/assets/", Static("/assets/", assets))
	srv.router.Handle(http.MethodGet, "/videos/", Static("/videos/", videos))

	srv.router.HandleFunc(http.MethodGet, "/stream", srv.stream)
	srv.router.HandleFunc(http.MethodGet, "/check", srv.check)
	srv.router.HandleFunc(http.MethodGet, "/all-videos/", srv.handleListVideos)
	srv.router.HandleFunc(http.MethodGet, "/", srv.handleIndex)
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.assets, "index.html"))
}

func (s *Server) handleListVideos(w http.ResponseWriter, r *http.Request) {
	var res struct {
		Items []Item `json:"items"`
	}
	res.Items = s.items.List()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("[ERROR] encondig response %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Frame struct {
	Frame  int    `json:"frame"`
	Total  int    `json:"total"`
	Millis int    `json:"millis"`
	Image  string `json:"image"`
}

type VideoData struct {
	Frame       int            `json:"frame,omitempty"`
	TotalFrames int            `json:"total_frames,omitempty"`
	Seconds     string         `json:"seconds,omitempty"`
	Complete    bool           `json:"complete,omitempty"`
	Faces       []facebox.Face `json:"faces,omitempty"`
	Thumbnail   *string        `json:"thumbnail,omitempty"`
}

func (s *Server) check(w http.ResponseWriter, r *http.Request) {
	var thumbnail *string
	// sent the headers for Server Side Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	enc := json.NewEncoder(w)

	// starts the video processing script
	filename := r.URL.Query().Get("name")
	flags := []string{"--path", path.Join(s.videos, filename), "--json", "True"}
	cmd := exec.CommandContext(r.Context(), "./video.py", flags...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("[ERROR] Getting the stdout pipe")
		return
	}
	cmd.Start()

	total := 0
	dec := json.NewDecoder(stdout)
	for {
		var f Frame
		err := dec.Decode(&f)
		if err == io.EOF {
			log.Println("[DEBUG] EOF", err)
			break
		}
		if err != nil {
			log.Println("[ERROR]", err)
			break
		}

		imgDec, err := base64.StdEncoding.DecodeString(f.Image)
		if err != nil {
			log.Printf("[ERROR] Error decoding the image %v\n", err)
			http.Error(w, "can not decode the image", http.StatusInternalServerError)
			return
		}
		faces, err := s.facebox.Check(bytes.NewReader(imgDec))
		total = f.Total

		thumbnail = nil
		for _, face := range faces {
			if face.Matched {
				thumbnail = &f.Image
			}
		}
		SendEvent(w, enc, VideoData{
			Frame:       f.Frame,
			TotalFrames: f.Total,
			Seconds:     (time.Duration(f.Millis/1000) * time.Second).String(),
			Complete:    false,
			Faces:       faces,
			Thumbnail:   thumbnail,
		})
	}
	cmd.Wait()
	SendEvent(w, enc, VideoData{
		Frame:       total,
		TotalFrames: total,
		Complete:    true,
	})
}

func SendEvent(w http.ResponseWriter, enc *json.Encoder, v interface{}) {
	w.Write([]byte("data: "))
	if err := enc.Encode(v); err != nil {
		log.Printf("[ERROR] Error encoding json %v\n", err)
		http.Error(w, "can not encode the json stream", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	const boundary = "informs"
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	filename := r.URL.Query().Get("name")
	flags := []string{"--path", path.Join(s.videos, filename)}
	cmd := exec.CommandContext(r.Context(), "./video.py", flags...)
	cmd.Stdout = w
	err := cmd.Run()
	if err != nil {
		log.Println("[ERROR] streaming the video", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Static gets a static file server for the specified path.
func Static(stripPrefix, dir string) http.Handler {
	h := http.StripPrefix(stripPrefix, http.FileServer(http.Dir(dir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
