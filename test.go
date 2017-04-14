package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"

    "log"
    "net/http"
    "net/url"
    "os"
    "os/user"
    "path/filepath"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
    cacheFile, err := tokenCacheFile()
    if err != nil {
        log.Fatalf("Unable to get path to cached credential file. %v", err)
    }
    tok, err := tokenFromFile(cacheFile)
    if err != nil {
        tok = getTokenFromWeb(config)
        saveToken(cacheFile, tok)
    }
    return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n", authURL)

    var code string
    if _, err := fmt.Scan(&code); err != nil {
        log.Fatalf("Unable to read authorization code %v", err)
    }

    tok, err := config.Exchange(oauth2.NoContext, code)
    if err != nil {
       log.Fatalf("Unable to retrieve token from web %v", err)
    }
    return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
    os.MkdirAll(tokenCacheDir, 0700)
    return filepath.Join(tokenCacheDir,
        url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    t := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(t)
    defer f.Close()
    return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", file)
    f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}

func showLable(usr string, srv *gmail.Service) {
    r, err := srv.Users.Labels.List(usr).Do()
    if err != nil {
        log.Fatalf("Unable to retrieve labels. %v", err)
    }
    if (len(r.Labels) > 0) {
        fmt.Print("Labels:\n")
        for _, l := range r.Labels {
            fmt.Printf("--- %s (%s)\n", l.Name, l.Id)
        }
        fmt.Println("\n", r.Labels[6], "\n")
    } else {
        fmt.Print("No labels found.")
    }
}

func showThreadsFromLable(usr string, srv *gmail.Service, lable_id string) {
    thr_list, err := srv.Users.Threads.List(usr).LabelIds(lable_id).Do()
    if err != nil {
      log.Fatalf("Unable to retrieve threads from labels %s. %v", lable_id, err)
    }
    //fmt.Printf("thr_list: %#v \n\n", thr_list)
    fmt.Printf("thr_list.Threads[1]: %#v \n\n", thr_list.Threads[0])
    //fmt.Println("Snippet: ", thr_list.Threads[1].Snippet, "\n")
    //fmt.Println("Id: ", thr_list.Threads[1].Id, "\n")
    thr_id := thr_list.Threads[0].Id
    thr_id2 := thr_list.Threads[2].Id
    thr, err := srv.Users.Threads.Get(usr, thr_id).Do()
    thr2, err := srv.Users.Threads.Get(usr, thr_id2).Do()
    fmt.Printf("thr: %#v \n\n", thr)
    fmt.Printf("thr2: %#v \n\n", thr2)
    
    for _, list1 := range thr_list.Threads {
        aaa := list1.Id
        fmt.Printf("--- %s\n", aaa)
        thr3, _ := srv.Users.Threads.Get(usr, aaa).Do()
        for _, list2 := range thr3.Messages[0].Payload.Headers {
            if (list2.Name == "Subject" || list2.Name == "Date") {
                fmt.Printf("--- %s: %s\n", list2.Name, list2.Value)
            }
        } 
    }
    
    fmt.Printf("thr2:  \n\n")
}

func main() {
    ctx := context.Background()

    b, err := ioutil.ReadFile("client_secret.json")
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
    }

    // If modifying these scopes, delete your previously saved credentials
    // at ~/.credentials/gmail-go-quickstart.json
    config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
    if err != nil {
        log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    client := getClient(ctx, config)

    usr := "me"
    srv, err := gmail.New(client)
    if err != nil {
        log.Fatalf("Unable to retrieve gmail Client %v", err)
    }
    
    showLable(usr, srv)

    showThreadsFromLable(usr, srv, "Label_10")
/*
    msg_list, err := srv.Users.Messages.List(usr).Do()
    fmt.Printf("thr: %#v \n\n", msg_list)

    msg_list, err := srv.Users.Messages.Get(usr).Do()

    thr, err := srv.Users.Threads.Get(usr, "15b65eb3467679fb").Do()
fmt.Printf("thr: %#v \n\n", thr)
*/
    fmt.Printf("Done \n\n")
}