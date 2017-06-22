package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"log"
	"net/http"
	"net/url"
	"net/mail"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"strconv"
	"encoding/csv"
	"time"
	

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	
)

type TargetMail struct {
	index int
	subject string
	//date string
	date time.Time
	stockId string
	op string
}

type TargetData struct {
	mail TargetMail
	symbol string
	open [12]int
	close [12]int
}

var targetMail []TargetMail
var targetData []TargetData

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
	if len(r.Labels) > 0 {
		fmt.Print("showLable:\n")
		for _, l := range r.Labels {
			fmt.Printf("--- %s (%s)\n", l.Name, l.Id)
		}
		fmt.Println("\n", r.Labels[6], "\n")
	} else {
		fmt.Print("No labels found.")
	}
}

func test(usr string, srv *gmail.Service, lableId string) {
	
	thrList := srv.Users.Threads.List(usr)
	thrList.LabelIds(lableId)
	thrList.MaxResults(1000)
	
	t, err := thrList.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve threads from labels %s. %v", lableId, err)
	}

	b, _ := t.MarshalJSON()
	os.Stdout.Write(b)

	for i, list1 := range t.Threads {
		aaa := list1.Id
		//fmt.Printf("--- %s\n", aaa)
		thr3, _ := srv.Users.Threads.Get(usr, aaa).Do()
		//buf := "--- " + strconv.FormatInt(int64(i), 10) + ":"
		var mMail TargetMail
		mMail.index = i
		for _, list2 := range thr3.Messages[0].Payload.Headers {
			switch list2.Name {
				case "Subject":
					mMail.subject = list2.Value
				case "Date":
					mMail.date, _ = mail.ParseDate(list2.Value)
			}
		}
		/*
		if (strings.Contains(list1.Snippet, "買進") ||
			strings.Contains(list1.Snippet, "持有")) {
			targetMail = append(targetMail, mMail)
		}
		*/
		targetMail = append(targetMail, mMail)
		fmt.Printf("--- %s: %s %s\n", strconv.FormatInt(int64(i), 10), mMail.subject, mMail.date)
		if i == 10 {
			//break
		}
	}

}

func showThreadsFromLable(usr string, srv *gmail.Service, lableId string) {
	thrList := srv.Users.Threads.List(usr)
	thrList.LabelIds(lableId)
	thrList.MaxResults(1000)
	thr, err := thrList.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve threads from labels %s. %v", lableId, err)
	}
	fmt.Printf("thr: %#v \n\n", thr)
	fmt.Printf("thr.Threads: len: %d, cap: %d \n\n", len(thr.Threads), cap(thr.Threads))
	fmt.Printf("thr.Threads[1]: %#v \n\n", thr.Threads[0])
	//fmt.Println("Snippet: ", thr_list.Threads[1].Snippet, "\n")
	//fmt.Println("Id: ", thr_list.Threads[1].Id, "\n")
	//thrId := thr.Threads[0].Id
	//thrId2 := thr.Threads[2].Id
	//thr, err := srv.Users.Threads.Get(usr, thrId).Do()
	//thr2, err := srv.Users.Threads.Get(usr, thrId2).Do()
	//fmt.Printf("thr: %#v \n\n", thr)
	//fmt.Printf("thr2: %#v \n\n", thr2)

	for i := len(thr.Threads)-1; i >= 0; i-- {
		id := thr.Threads[i].Id

		t, _ := srv.Users.Threads.Get(usr, id).Do()
		//fmt.Printf("t: %#v \n\n", t)
		//buf := "--- " + strconv.FormatInt(int64(i), 10) + ":"
		var mMail TargetMail
		mMail.index = i
		for _, list2 := range t.Messages[0].Payload.Headers {
			switch list2.Name {
				case "Subject":
					mMail.subject = list2.Value
				case "Date":
					mMail.date, _ = mail.ParseDate(list2.Value)
			}
		}
		ScanStockId(&mMail)

		targetMail = append(targetMail, mMail)

		fmt.Printf("--- %s: %s\n", strconv.FormatInt(int64(i), 10), mMail.subject)
		if i == 350 {
			break
		}
	}
}

func writeToCsv() {
	file := openFile("mail.csv", os.O_WRONLY|os.O_CREATE)
	w := csv.NewWriter(file.writer)
	for _, v := range targetMail {
		t := fmt.Sprintf("%d/%d/%d", v.date.Year(), v.date.Month(), v.date.Day())
		s := []string{v.stockId, t, v.op, v.subject}
		w.Write(s)
	}
	w.Flush()
}

func ScanStockId(m *TargetMail) {
		found := false
		if strings.Contains(m.subject, "買進") {
			m.op = "買進"
			found = true
		} else if strings.Contains(m.subject, "持有") {
			m.op = "持有"
			found = true
		}
		if found {
			re := regexp.MustCompile("\\(\\d\\d\\d\\d")
			s := re.FindString(m.subject)
			if s != "" {
				m.stockId = s[1:]
				//fmt.Printf("--- s: %s, %s\n", m.stockId, m.subject)
			}
		}
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
	//test(usr, srv, "Label_10")
	//showLable(usr, srv)
	showThreadsFromLable(usr, srv, "Label_10")
	writeToCsv()
	
}
