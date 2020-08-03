package main

import (
    "os"
    "sync"
    "encoding/json"
    "fmt"
    "net"
    "net/http"
    "strconv"
    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
    "io/ioutil"
    "time"
    "strings"
    "text/template"
    "net/smtp"
    "bytes"
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

type ProxyInfo struct {
    Name          string     `json:"name"`
    ContainerName string     `json:"container_name"`
    ProxyType     string     `json:"proxy_type"`
    RemotePort    int        `json:"remote_port"`
    LocalPort     int        `json:"local_port"`
    Email         string     `json:"email"`
    ClientPrefix  string     `json:"frps_prefix"`
    Url           string     `json:"url"`
    Active        bool       `json:"active"`
    Notified      bool       `json:"notified"`
    
}

type ProxyList struct {
    Proxies map[string]ProxyInfo `json:"proxies"`
}

type DisplayProxyList struct {
    Active      map[string][]ProxyInfo    `json:"active"`
    Inactive    map[string][]ProxyInfo    `json:"inactive"`
}

var mutex sync.RWMutex
var references = ProxyList{Proxies: make(map[string]ProxyInfo)}

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
func saveProxyLinksJSON() {
    file, _ := json.MarshalIndent(references, "", " ")
 
    _ = ioutil.WriteFile("links.json", file, 0644)
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
                
                metas, ok := r.Content["metas"]
            
                // do nothing if metas does not exists                
                if ok && metas != nil {
                    
                    metas := metas.(map[string]interface{})
                    
                    notify_email, ok_email := metas["notify_email"]
                    frpc_prefix, ok_prefix := metas["frpc_prefix"]
                    local_port, ok_port := metas["local_port"]
                    
                    // do nothing of there is no notify_email, frpc_prefix or local_port in metas
                    
                    if ok_email && ok_prefix && ok_port {
                        
                        var key = fmt.Sprintf("%v:%v", r.Content["proxy_name"], r.Content["proxy_type"])
                        
                        var url string = getEnv("FRPS_SUBDOMAIN_HOST","example.com")
                        var remote_port int = 0
                        
                        if r.Content["proxy_type"] == "tcp" || r.Content["proxy_type"] == "udp" {
                            remote_port = int(r.Content["remote_port"].(float64))
                            
                            url = fmt.Sprintf("%v:%v", url, remote_port)
                            
                        } else if r.Content["proxy_type"] == "http" {
                            remote_port = 80
                            
                            url =  fmt.Sprintf("http://%v.%v", r.Content["subdomain"], url)
                            
                        }  else if r.Content["proxy_type"] == "https" {
                            remote_port = 443
                            
                            url =  fmt.Sprintf("https://%v.%v", r.Content["subdomain"], url)
                            
                        }
                        
                        local_port, _ := strconv.Atoi(local_port.(string))
                        
                        var container_name string = r.Content["proxy_name"].(string)
                        
                        // to get actual container name by removing prefix name and port number suffix
                        container_name = strings.Replace(container_name, fmt.Sprintf("%s_",frpc_prefix),"", 1)
                        container_name = strings.Replace(container_name, fmt.Sprintf("_%d",local_port), "", 1)
                        
                        ref := ProxyInfo{ Name: key, 
                                          ContainerName: container_name,
                                          ProxyType: r.Content["proxy_type"].(string),
                                          RemotePort: remote_port,
                                          LocalPort: int(local_port),
                                          Email: notify_email.(string),
                                          ClientPrefix: frpc_prefix.(string),
                                          Url: url,
                                          Active: true,
                                          Notified: false }
                        
                        mutex.RLock()

                        link, ok := references.Proxies[key]

                        if !ok || !cmp.Equal(link, ref, cmpopts.IgnoreFields(ProxyInfo{}, "Notified")) {
                            // update reference if 
                            //   - does not exists at all
                            //   - already exsits but is not the same
                            references.Proxies[key] = ref
                            
                            // save new links
                            saveProxyLinksJSON()
                            
                        }
                        
                        mutex.RUnlock()
                    }
                }
                
                o.Reject = false
                o.Unchange = true
                
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

func notifier_main() {
    
    //var frps_subdomain_host string = getEnv("FRPS_SUBDOMAIN_HOST","example.com")
    
    // infinite loop that 
    //  - checks for changes of the links.json
    //  - if at least 30 sec since last change 
    var last_notified = time.Now()
    
    var FRPS_LINK_NOTIFIER_DELAY_SEC int = getEnvInt("FRPS_LINK_NOTIFIER_DELAY_SEC", 15)
    var FRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC int = getEnvInt("FRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC", 5)
    var FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC int = getEnvInt("FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC", 2)
    
    var FRPS_LINK_NOTIFIER_EMAIL_SUBJECT string = getEnv("FRPS_LINK_NOTIFIER_EMAIL_SUBJECT", "Reverse proxy links update")
    
    var FRPS_LINK_NOTIFIER_SMTP_ACCOUNT string = getEnv("FRPS_LINK_NOTIFIER_SMTP_ACCOUNT", "")
    var FRPS_LINK_NOTIFIER_SMTP_PASS string = getEnv("FRPS_LINK_NOTIFIER_SMTP_PASS", "")    
    var FRPS_LINK_NOTIFIER_SMTP_SERVER string = getEnv("FRPS_LINK_NOTIFIER_SMTP_SERVER", "")    
                
    tpl, err := template.ParseFiles("notification_email.html.tpl")
    
    if err != nil {
        fmt.Println("ERROR in notifier_main(): missing notification_email.html.tpl template file. E-mail notification will not be performed !!")
        return
    }
    
    auth := smtp.PlainAuth("", FRPS_LINK_NOTIFIER_SMTP_ACCOUNT, FRPS_LINK_NOTIFIER_SMTP_PASS, strings.Split(FRPS_LINK_NOTIFIER_SMTP_SERVER, ":")[0])
    
    if auth == nil {
        fmt.Printf("ERROR in notifier_main(): server authentication failed (%s)\n", FRPS_LINK_NOTIFIER_SMTP_SERVER)
        return
    }
        
    fmt.Printf("In notifier_main(): started notification loop with:\n" +
                "\tFRPS_LINK_NOTIFIER_DELAY_SEC=%d\n" +
                "\tFRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC=%d\n" +
                "\tFRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC=%d\n" +
                "\tFRPS_LINK_NOTIFIER_SMTP_SERVER=%s\n" +
                "\tFRPS_LINK_NOTIFIER_SMTP_ACCOUNT=%s\n", 
                FRPS_LINK_NOTIFIER_DELAY_SEC, FRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC, FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC, FRPS_LINK_NOTIFIER_SMTP_SERVER, FRPS_LINK_NOTIFIER_SMTP_ACCOUNT)
    
    for {
        file, err := os.Stat("links.json")
        if err == nil {            

            modified_time := file.ModTime()
            
            if modified_time.After(last_notified) && time.Now().After(modified_time.Add(time.Duration(FRPS_LINK_NOTIFIER_DELAY_SEC) * time.Second)) {
                
                
                fmt.Printf("In notifier_main(): at least %d sec since last modification .. doing notification now\n", FRPS_LINK_NOTIFIER_DELAY_SEC)
                
                mutex.RLock()
                
                // first check for validity of each connection and flag unresponsive ones
                should_notify := false
                num_active := 0
                
                for name, _ := range references.Proxies {
                    proxy_ref := references.Proxies[name]
                    
                    // check if connection is active 
                    proxy_ref.Active = check_connection(proxy_ref, FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC)
                    
                    if proxy_ref.Active {
                        num_active = num_active + 1
                    }
                    
                    // should notify only if any proxy is active and has not been yet notified
                    should_notify = should_notify || (proxy_ref.Active && !proxy_ref.Notified)
                                        
                    // update value                    
                    references.Proxies[name] = proxy_ref
                    
                }
                
                // save updated links
                saveProxyLinksJSON()
                
                // then group references by email notifications 
                var gruped_proxies = make(map[string][]ProxyInfo)
                for _, proxy_ref := range references.Proxies {
                    gruped_proxies[proxy_ref.Email] = append(gruped_proxies[proxy_ref.Email], proxy_ref)
                }
                
                // perform user notification if needed
                if should_notify {
                    
                    var num_sent_emails int = 0
                    
                    // go over each group and create a notification list                    
                    for email, proxy_ref_list := range gruped_proxies {
                        
                        should_notify := false
                        
                        // first chech if any of the users connections have not been yet notified
                        for _, proxy_ref := range proxy_ref_list {
                            should_notify = should_notify || ! proxy_ref.Notified
                        }
                        
                        // do notification only if user has not been notified for at least one connection
                        if should_notify {
                            
                            var display_proxy_list = DisplayProxyList{Active: make(map[string][]ProxyInfo), 
                                                                      Inactive: make(map[string][]ProxyInfo)}
                            
                            // get active connections first 
                            for _, proxy_ref := range proxy_ref_list {
                                if proxy_ref.Active {
                                    // copy to list of active proxies for notification mail
                                    display_proxy_list.Active[proxy_ref.ContainerName] = append(display_proxy_list.Active[proxy_ref.Name], proxy_ref)
                                    
                                    // mark as notified 
                                    var actual_ref = references.Proxies[proxy_ref.Name]
                                    actual_ref.Notified = true
                                    
                                    references.Proxies[proxy_ref.Name] = actual_ref
                                    
                                }
                            }
                            
                            // get inactive connections last
                            for _, proxy_ref := range proxy_ref_list {
                                if !proxy_ref.Active {
                                    // copy to list of active proxies for notification mail
                                    display_proxy_list.Inactive[proxy_ref.ContainerName] = append(display_proxy_list.Inactive[proxy_ref.Name], proxy_ref)
                                    
                                    // mark as notified 
                                    var actual_ref = references.Proxies[proxy_ref.Name]
                                    actual_ref.Notified = true
                                    
                                    references.Proxies[proxy_ref.Name] = actual_ref

                                }
                            }
                            
                            
                            var msg bytes.Buffer
                            err = tpl.Execute(&msg, display_proxy_list)
                            
                            if err != nil {
                                fmt.Println(err)
                                return
                            }
                            
                            var msg_str string = fmt.Sprintf("To: %s\r\n", email) +
                                                 fmt.Sprintf("Subject: %s\r\n", FRPS_LINK_NOTIFIER_EMAIL_SUBJECT ) +
                                                 "\r\n" +
                                                 fmt.Sprintf("%s\r\n",msg.String())

                            err := smtp.SendMail(FRPS_LINK_NOTIFIER_SMTP_SERVER, auth, FRPS_LINK_NOTIFIER_SMTP_ACCOUNT, []string{email}, []byte(msg_str))
                            
                            if err != nil {
                                fmt.Printf("ERROR in notifier_main(): when sending mail to %s got '%s'\n", email, err)
                                continue
                            }
                            
                            num_sent_emails = num_sent_emails + 1
                        }                       
                        
                    }
                    
                    fmt.Printf("In notifier_main(): notification email sent to %d recipient(s)\n", num_sent_emails)
                    
                    // save updated links
                    saveProxyLinksJSON()
                }
                
                
                mutex.RUnlock()
                last_notified = time.Now()            
            }
        }
        time.Sleep(time.Duration(FRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC) * time.Second)
	}
    
}

func check_connection(proxy_ref ProxyInfo, FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC int) bool {
    var ok bool = false
    if proxy_ref.ProxyType == "tcp" || proxy_ref.ProxyType == "udp" {        
        // check using direct connection                    
        conn, err := net.DialTimeout(proxy_ref.ProxyType, proxy_ref.Url, time.Duration(FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC)*time.Second)
        if err == nil && conn != nil  {
            defer conn.Close()  

            // connection is valid so we retain it
            ok = true
        }
        
    } else if proxy_ref.ProxyType == "http" || proxy_ref.ProxyType == "https" {
        
        client := http.Client{Timeout: time.Duration(FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC) * time.Second}        
        // check using HTTP request                    
        _, err := client.Get(proxy_ref.Url)
        if err == nil  {
            // connection is valid so we retain it
            ok = true
        } 
    }
    
    return ok
}


func main() {

    file, err := ioutil.ReadFile("links.json")
    
    if err == nil {    
        err = json.Unmarshal([]byte(file), &references)
    }

    go notifier_main()
    
    http.HandleFunc("/", handler)
    http.ListenAndServe(fmt.Sprintf(":%d", getEnvInt("PLUGIN_PORT", 9003)), nil)
}

