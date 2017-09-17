package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"bytes"
)

var (
	flagUrl = flag.String("host", "http://localhost:15672", "Management API root url")
	flagVhost = flag.String("vhost", "/", "Virtual host")
	flagPattern= flag.String("queue", ".+", "Check queue pattern")
	flagExclude = flag.String("exclude", "", "Exclude queue pattern")
	flagErr = flag.Int("error", 3, "Error message threshold")
	flagWarn = flag.Int("warn", 2, "Warning message threshold")
	flagUser = flag.String("user", "guest", "Username")
	flagPassword = flag.String("password", "guest", "Password")
)

const pathApiQueues = "/api/queues/"

type Queue struct {
	Name           string
	Messages_Ready int
	Consumers      int
}

func (q Queue) Println() {
	fmt.Printf("%+v\n", q)
}

func (q Queue) ToString() string {
	return fmt.Sprintf("%+v\n", q)
}

func Fatal(i ...interface{}) {
	fmt.Println(i...)
	os.Exit(3)
}

func Get(url *url.URL, username string, password string) *http.Response {
	client := &http.Client{}
	r, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		Fatal(err)
	}
	r.SetBasicAuth(username, password)
	resp, err := client.Do(r)
	if err != nil {
		Fatal(err)
	}
	return resp
}

func LoadQueues(resp *http.Response) []Queue {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Fatal(err)
	}
	var queues []Queue
	err = json.Unmarshal(body, &queues)
	if err != nil {
		Fatal(err)
	}
	return queues
}

func UrlJoin(base string, path string) *url.URL {
	b, err := url.Parse(base)
	if err != nil {
		Fatal(err)
	}
	var resp_url *url.URL
	p, err := url.Parse(path)
	if err != nil {
		Fatal(err)
	}
	resp_url = b.ResolveReference(p)
	if err != nil {
		Fatal(err)
	}
	return resp_url
}

func CheckQueue(q Queue, warn int, err int) int {
	if q.Messages_Ready >= err {
		return 2
	} else if q.Messages_Ready >= warn {
		return 1
	}
	return 0
}

func Check(queues []Queue, pattern string, exclude string, warn int, err int) (int, string) {
	exitCode := 0
	var buf bytes.Buffer
	r := regexp.MustCompilePOSIX(pattern)
	e := regexp.MustCompilePOSIX(exclude)
	for _, q := range queues {
		if exclude != "" && e.MatchString(q.Name) == true {
			continue
		}
		if r.MatchString(q.Name) == true {
			r := CheckQueue(q, warn, err)
			if r > 0 {
				buf.WriteString(q.ToString())
			}
			if r > exitCode {
				exitCode = r
			}
		}
	}
	if buf.Len() == 0 {
		buf.WriteString("OK\n")
	}
	return exitCode, buf.String()
}

func main() {
	flag.Parse()

	apiUrl := UrlJoin(*flagUrl, pathApiQueues + url.QueryEscape(*flagVhost))
	resp := Get(apiUrl, *flagUser, *flagPassword)

	if resp.StatusCode != 200 {
		Fatal(apiUrl, "-", resp.Status)
	}

	queues := LoadQueues(resp)
	exitCode, out := Check(queues,
		*flagPattern,
		*flagExclude,
		*flagWarn,
		*flagErr)
	fmt.Print(out)
	os.Exit(exitCode)
}
