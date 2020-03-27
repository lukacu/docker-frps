package main

import (
    "os"
    "sync"
    "encoding/json"
    "fmt"
    "net/http"
    "bufio"
    "regexp"
    "strconv"
)

type Request struct {
	Version string      `json:"version"`
	Op      string      `json:"op"`
	Content map[string]interface{} `json:"content"`
}

type Response struct {
	Reject       bool        `json:"reject"`
	RejectReason string      `json:"reject_reason"`
	Unchange     bool        `json:"unchange"`
	Content      map[string]interface{} `json:"content"`
}

var mutex sync.RWMutex
var ports = make(map[string]int)

func getEnv(name string, def string) string {
    val := os.Getenv(name)

    if val == "" {
        return def
    }

    return val

}

func getEnvInt(name string, def int) int {
    val := os.Getenv(name)

    if val == "" {
        return def
    }

    ival, _ := strconv.ParseInt(val, 10, 32)

    return int(ival)
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var portMin int = getEnvInt("PLUGIN_PORT_MIN", 30000)
var portMax int = getEnvInt("PLUGIN_PORT_MAX", 30900)

func savePortMapping() {

    f, err := os.Create("ports.map")
    check(err)

    for k, v := range ports {
        _, err := f.WriteString(fmt.Sprintf("%v %v\n", k, v))
        check(err)
    }

    defer f.Close()
}

func handler(w http.ResponseWriter, r *http.Request) {

        switch r.Method {
        case "POST":
                d := json.NewDecoder(r.Body)
                r := &Request{}
                o := &Response{}
                err := d.Decode(r)
                if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                }

                if r.Op != "NewProxy" {
                    w.WriteHeader(http.StatusMethodNotAllowed)
                    fmt.Fprintf(w, "Not allowed.")
                    return
                }

                if r.Content["proxy_type"] == "tcp" || r.Content["proxy_type"] == "udp" {

                    var key = fmt.Sprintf("%v:%v", r.Content["proxy_name"], r.Content["proxy_type"])
                    var port int = int(r.Content["remote_port"].(float64))

                    // Allocate or retrieve port
                    if port == 0 {

                        mutex.RLock()

                        port, ok := ports[key]

                        if !ok {

                            var allocated = make(map[int]bool)

                            for _, v := range ports {
	                            allocated[v] = true
                            }

                            for i := portMin; i <= portMax; i++ {

                                if !allocated[i] {
                                    port = i
                                    break
                                }

                            }

                            if port == 0 {
                                o.Reject = true
                                o.RejectReason = "All available ports already taken"

                            } else {

                                ports[key] = port
                                savePortMapping()

                                o.Reject = false
                                o.Unchange = false
                                o.Content = r.Content
                                o.Content["remote_port"] = port

                            }

                        } else {

                            o.Reject = false
                            o.Unchange = false
                            o.Content = r.Content
                            o.Content["remote_port"] = port

                        }

                        mutex.RUnlock()

                    } else {
                        // Verify that port is not taken

                        mutex.RLock()

                        var found bool = false

                        for k, v := range ports {
                            if v == port {
                                if k == key {
                                    o.Reject = false
                                    o.Unchange = true
                                } else {
                                    o.Reject = true
                                    o.RejectReason = "Port already taken by another proxy"
                                }
                                found = true
                            }
                        }

                        if !found {
                            if port >= portMin && port <= portMax {
                                ports[key] = port
                                o.Reject = false
                                o.Unchange = true

                                savePortMapping()

                            } else {
                                o.Reject = true
                                o.RejectReason = "Illegal port number"
                            }
                        }

                        mutex.RUnlock()

                    }

                } else {
                    o.Reject = false
                    o.Unchange = true
                    return

                }

                js, err := json.Marshal(o)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(js)

        default:
                w.WriteHeader(http.StatusMethodNotAllowed)
                fmt.Fprintf(w, "I can't do that.")
        }
}

func main() {

    f, err := os.Open("ports.map")

    if !os.IsNotExist(err) {
        var lineParser = regexp.MustCompile(`^([^\\w]+) ([0-9]+)$`)

        s := bufio.NewScanner(f)
        for s.Scan() {
            line := s.Text()

            matches := lineParser.FindSubmatch([]byte(line))

            if len(matches) == 3 {
                port64, _ := strconv.ParseInt(string(matches[2]), 10, 32)
                port := int(port64)

                if port >= portMin && port <= portMax {
                    ports[string(matches[1])] = int(port)
                }

            }

        }


        f.Close()
    }

    http.HandleFunc("/", handler)
    http.ListenAndServe(fmt.Sprintf(":%d", getEnvInt("PLUGIN_PORT", 9001)), nil)
}

