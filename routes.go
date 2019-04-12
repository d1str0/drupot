package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

var templates = template.Must(template.ParseGlob("templates/*"))

var (
	ErrDrupal422 = errors.New("The website encountered an unexpected error. Please try again later.")
)

type Page struct {
	Title string
	Host  string
}

func drupal404Handler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "drupal-404.html", getHost(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// staticHandler provides static pages depending on the request. If
// CHANGELOG.txt is requested, return the appropriate Changelog file and flag
// the IP. Otherwise, return the index page and check whether to record the
// http.Request.
func drupalIndexHandler(app App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		if path == "CHANGELOG.txt" {
			if app.Config.Drupal.ChangelogEnabled {
				http.ServeFile(w, r, app.Config.Drupal.ChangelogFilepath)
			} else {
				// Changelog files in Drupal 8.* are generally not served
				drupal404Handler(w, r)
			}
			saveIP(app, r)
		} else {
			checkIP(app, r)
			filename := fmt.Sprintf("index-%s.html", app.Config.Drupal.Version)
			host := fmt.Sprintf("http://%s", app.SensorIP)
			p := Page{
				Title: app.Config.Drupal.SiteName,
				Host:  host,
			}
			fmt.Println(filename)
			err := templates.ExecuteTemplate(w, filename, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func drupalNodeHandler(app App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		saveIP(app, r)
		http.Error(w, ErrDrupal422.Error(), http.StatusUnprocessableEntity)
	}
}

func fileServe(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("public%s", r.URL.Path)
	http.ServeFile(w, r, path)
}

// routes sets up the necessary http routing for the webapp.
func routes(app App) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", drupalIndexHandler(app))
	mux.HandleFunc("/core/", fileServe)
	mux.HandleFunc("/sites/", fileServe)
	mux.HandleFunc("/logo.svg", fileServe)
	mux.HandleFunc("/node/", drupalNodeHandler(app))
	return mux
}

// saveIP flags the given IP so that if we see it in the future we can record
// its requests.
func saveIP(app App, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	// If this is a previously unseen IP, let's remember them.
	if !app.SeenIP[ip] {
		recordRequest(app, r, false)
		app.SeenIPLock.Lock()
		defer app.SeenIPLock.Unlock()

		app.SeenIP[ip] = true
		fmt.Printf("New CHANGELOG request: %s, %s\n", ip, r.URL.Path)
	}
}

// checkIP checks to see if this IP has been flagged before. If so, we
// record the http.Request.
func checkIP(app App, r *http.Request) {
	// If this is a previously seen IP, let's record what they do.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	app.SeenIPLock.RLock()
	defer app.SeenIPLock.RUnlock()

	// If we saw this IP request our CHANGELOG, record whatever they do next.
	if app.SeenIP[ip] {
		recordRequest(app, r, true)
		fmt.Printf("Seen request from: %s, %s\n", ip, r.URL.Path)
	} else {
		recordRequest(app, r, false)
		fmt.Printf("Seen request from: %s, %s\n", ip, r.URL.Path)
	}
}

// Msg normalizes the recieved request and allows for easy marshaling into JSON.
type Msg struct {
	Protocol      string
	App           string
	Channel       string
	Sensor        string
	DestPort      int
	DestIp        string
	SrcPort       int
	SrcIp         string
	Meta          string
	Signature     string
	Fingerprinted bool
	Request       *RequestJson
}

// recordRequest will parse the http.Request and put it into a normalized format
// and then marshal to JSON. This can then be sent on an hpfeeds channel or
// logged to a file directly.
func recordRequest(app App, r *http.Request, fingerprinted bool) {
	ip, p, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	rj := TrimRequest(r)

	// Populate data to send
	pub_msg := Msg{
		Protocol:      r.Proto,
		App:           "Drupot",
		Channel:       app.Config.Hpfeeds.Channel,
		Sensor:        app.SensorUUID.String(),
		DestPort:      app.Config.Drupal.Port,
		DestIp:        app.SensorIP,
		SrcPort:       port,
		SrcIp:         ip,
		Meta:          app.Config.Hpfeeds.Meta,
		Fingerprinted: fingerprinted,
		Request:       rj,
	}

	// Marshal it to json so we can send it over hpfeeds.
	buf, err := json.Marshal(pub_msg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// Send to hpfeeds broker
	if app.Config.Hpfeeds.Enabled {
		app.Publish <- buf
	}
}

func TrimRequest(r *http.Request) *RequestJson {
	body, _ := ioutil.ReadAll(r.Body)
	r.ParseForm()
	rj := &RequestJson{
		Method:           r.Method,
		URL:              r.URL,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		Body:             body,
		TransferEncoding: r.TransferEncoding,
		Host:             r.Host,
		PostForm:         r.PostForm,
	}

	return rj
}

type RequestJson struct {
	Method           string
	URL              *url.URL
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             []byte
	TransferEncoding []string
	Host             string
	PostForm         url.Values
}

// getHost tries its best to return the request host.
func getHost(r *http.Request) string {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host

	return r.URL.String()
}
