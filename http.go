package main

import (
    "net/http"
    "log"
    "io/ioutil"
    "time"
    "encoding/json"
    "github.com/julienschmidt/httprouter"
    "github.com/unrolled/render"
)

type GoLink struct {
    Url string
    Key string
    Count int
    Created int32
    LastUsed int32
}

var lnks map[string]GoLink

/* 
	If no key was given show a list of available links
*/
func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ren := render.New(render.Options{
		Layout: "layout",
	})
	ren.HTML(w, http.StatusOK, "index", lnks)
}

/* 
	Used to redirect the actual request (if it exists) otherwise show an add page 
*/
func redirectHandle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Print("Got a redirect request for '" + key + "'")

	goLink, ok := lnks[key]
	if ok {
		url :=  goLink.Url
		updateGoLink(goLink)
    	http.Redirect(w, r, url, 302)
    	return
	}

    // If we didn't find a match assume we want to create a new one
    url := "/" + key + "/add"
    http.Redirect(w, r, url, 302)
}

/* 
	Update an existing go link
*/
func updateGoLink(goLink GoLink) {
	now := int32(time.Now().Unix())
	goLink.LastUsed = now
	goLink.Count = goLink.Count + 1
	lnks[goLink.Key] = goLink
	saveLinks()
}

/* 
	Add/Edit a new link 
*/
func addLinkHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if r.URL.Query().Get("key") != "" {
		key = r.URL.Query().Get("key")
	}
	url := r.URL.Query().Get("url")
	addBtn := r.URL.Query().Get("add_btn")
	now := int32(time.Now().Unix())
	ren := render.New(render.Options{
		Layout: "layout",
	})

	// If both key and url are passed in add it to the db.
	if (addBtn != "") {
		newLink := GoLink{Url: url, Key: key, LastUsed: now, Created: now, Count: 0}
		lnks[key] = newLink
		saveLinks()
		ren.HTML(w, http.StatusOK, "add_complete", newLink)
	} else {
		goLink, ok := lnks[key]
		if !ok {
			goLink = GoLink{Url: url, Key: key, LastUsed: 0, Created: 0, Count: 0}
		}
		ren.HTML(w, http.StatusOK, "add_form", goLink)
	}
}

/* 
	Delete an existing go link
*/
func deleteLinkHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	delete(lnks, key)
	saveLinks()
	http.Redirect(w, r, "/?delete=" + key, 302)
}

/*
	Save the links in memory to disk
*/
func saveLinks() {
	b, _ := json.Marshal(lnks)
    err := ioutil.WriteFile("links.json", b, 0644)
    log.Print(err)
}

/*
	Read the links from disk and store them in memory
*/
func readLinks() {
	dat, err := ioutil.ReadFile("links.json")
	log.Print(err)
    json.Unmarshal(dat, &lnks)
}

func main() {
	// init the url list from disk
	lnks = make(map[string]GoLink)

	readLinks()

	// Setup routes
	router := httprouter.New()
    router.GET("/", indexHandler)
    router.GET("/:key", redirectHandle)
    router.GET("/:key/add", addLinkHandler)
    router.GET("/:key/delete", deleteLinkHandler)

    log.Fatal(http.ListenAndServe(":8080", router))
}

