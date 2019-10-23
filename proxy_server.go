package main

import (
    "net/http"
    "io"
    "fmt"
    "os"
    "bufio"
)

type Proxy struct {
    // use a map for constant time lookup
    BlockedSites map[string]string
}

func CreateProxy() *Proxy {
    rv := new(Proxy)
    rv.BlockedSites = make(map[string]string)
    return rv
}

func (p *Proxy) ReadConfig (path string) {
    file, err := os.Open(path)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        site := scanner.Text()
        p.BlockedSites[site] = site
    }
}

func (p *Proxy) StartServer() {
    http.HandleFunc("/", p.HandleRequest)
    http.ListenAndServe(":8080", nil)
}

func (p *Proxy) HandleRequest(w http.ResponseWriter, r *http.Request) {

    // check if this site is blocked
    _, blocked := p.BlockedSites[r.URL.Host]
    if blocked {
        fmt.Fprint(w, "Page blocked!")
        return
    }

    // format new request
    request_path := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)

    // create new HTTP request with the target URL (everything else is the same)
    new_request, err := http.NewRequest(r.Method, request_path, r.Body)

    // send request to server
    fmt.Printf("Sending %s request to %s\n", r.Method, request_path)
    client := &http.Client{}
    res, err := client.Do(new_request)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer res.Body.Close()

    // copy the headers over to the ResponseWriter. 
    //res.Header is a map of string -> slice (string)
    for key, slice := range res.Header {
        for _, val := range slice {
            w.Header().Add(key, val)
        }
    }

    // forward response to client
    _, err = io.Copy(w, res.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
}

func main() {
    p := CreateProxy()
    p.ReadConfig("blocked_sites.txt")
    p.StartServer()
}
