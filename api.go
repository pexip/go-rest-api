package pexrest

import (
    "net"
    "net/http"
    "io/ioutil"
    "fmt"
    "encoding/json"
    "os"
    "bytes"
    "bufio"
    "strconv"
    "time"
    "crypto/tls"
    "strings"
)

type Event struct {
    name string
    id string
    data string
}

type Context struct {
    Host string
    Pin string
    Conference string
    DisplayName string
    UUID string
    Expires int
    Destination string

    token string
    timer *time.Timer
}

func NewContext(display_name string, destination string) (ctx * Context) {
    ctx = new(Context)
    ctx.DisplayName = display_name
    ctx.setDestination(destination)
    return ctx
}

func (ctx * Context) setDestination(uri string) (succeeded bool) {
    confhost := strings.Split(uri, "@")
    if len(confhost) != 2 {
        return false
    }
    ctx.Conference = confhost[0]
    host := confhost[1]
    _, addrs, err := net.LookupSRV("pexapp", "tcp", host)
    if err == nil && len(addrs) > 0 {
        ctx.Destination = fmt.Sprintf("%s:%d", addrs[0].Target[:len(addrs[0].Target)-1], addrs[0].Port)
    } else {
        ctx.Destination = host
    }
    fmt.Fprintf(os.Stdout, "rest: connecting to %s\n", ctx.Destination)
    return true
}

func (ctx * Context) Post(path string, postBody []byte) (bool, []byte) {
    url := fmt.Sprintf("https://%s%s", ctx.Destination, path)
    transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}
    client := &http.Client{Transport: transport}
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(postBody))
    if len(ctx.token) > 0 {
        req.Header.Add("token", ctx.token)
    } else {
        req.Header.Add("pin", ctx.Pin)
    }
    res, err := client.Do(req)
    if err != nil {
        fmt.Fprintf(os.Stdout, "rest: failed to connect: %s\n", err)
        return false, nil
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        fmt.Fprintf(os.Stdout, "rest: failed to connect: %s\n", res.Status)
        return false, nil
    }
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return false, body
    }

    return true, body
}

func readEventLine(r * bufio.Reader) (name string, value string) {
    line, _, err := r.ReadLine()
    if err != nil {
        panic(err)
    }
    if len(line) == 0 {
        return readEventLine(r)
    }
    tokens := strings.Split(string(line), ": ")
    return tokens[0], tokens[1]
}

func getEvent(r * bufio.Reader) (event Event) {
    field_name, field_value := readEventLine(r)
    if len(field_name) > 0 {
        if (field_name != "event") { panic("event has no name"); }
        event.name = field_value
        field_name, field_value = readEventLine(r)
        if (field_name == "id") {
            event.id = field_value
            field_name, field_value = readEventLine(r)
        }
        if (field_name != "data") { panic("event has no data"); }
        event.data = field_value
    } else {
        event.name = field_value
    }
    return event
}

func (ctx * Context) listenOnEvents(path string) (succeeded bool) {
    succeeded = true
    url := fmt.Sprintf("https://%s%s", ctx.Destination, path)
    transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}
    client := &http.Client{Transport: transport}
    req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
    req.Header.Add("token", ctx.token)
    req.Header.Add("Accept", "text/event-stream")
    res, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        fmt.Fprintf(os.Stdout, "rest: failed to connect: %s\n", res.Status)
        succeeded = false
    }
    reader := bufio.NewReaderSize(res.Body, 4096)

    for {
        e := getEvent(reader)
        fmt.Fprintf(os.Stdout, "rest: got event %+v\n", e)
        if e.name == "bye" { break }
    }

    return succeeded
}

func (ctx * Context) Get(path string) (succeeded bool, body []byte) {
    succeeded = true
    url := fmt.Sprintf("https://%s%s", ctx.Destination, path)
    transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}
    client := &http.Client{Transport: transport}
    req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
    req.Header.Add("token", ctx.token)
    res, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        fmt.Fprintf(os.Stdout, "rest: failed to connect: %s\n", res.Status)
        succeeded = false
    }
    body, err = ioutil.ReadAll(res.Body)
    if err != nil {
        panic(err)
    }

    return succeeded, body
}

func (ctx * Context) GetParticipants() (result ParticipantsResult) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/participants", ctx.Conference)
    _, body := ctx.Get(url)
    err := json.Unmarshal (body, &result)
    if err != nil {
        panic(err)
    }
    return result
}

func (ctx * Context) GetConference() (conferences []byte) {
    return conferences
}


func (ctx * Context) GetStatistics(participant Participant) (succeeded bool, results StatisticsResult) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/participants/%s/statistics", ctx.Conference, participant.UUID)
    succeeded, body := ctx.Get(url)
    if !succeeded {
        return false, results
    }
    err := json.Unmarshal (body, &results)
    if err != nil {
        panic(err)
    }
    return true, results
}

func (ctx * Context) ReleaseToken() bool {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/release_token", ctx.Conference)
    _, body := ctx.Post(url, nil)
    var result GenericResult
    err := json.Unmarshal (body, &result)
    if err != nil {
        panic(err)
    }
    if ctx.timer != nil {
        ctx.timer.Stop()
    }
    return result.Status == "success"
}


func (ctx * Context) loginAgain() {
    for {
        ctx.timer = time.NewTimer(time.Second * time.Duration(ctx.Expires/4))
        <- ctx.timer.C
        if succeeded := ctx.RefreshToken(); !succeeded {
            continue
        }
    }
}

func (ctx * Context) RequestToken() (bool) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/request_token?display_name=%s", ctx.Conference, ctx.DisplayName)
    succeeded, body := ctx.Post(url, nil)
    if !succeeded {
        return false
    }
    var token GenericTokenResult
    err := json.Unmarshal (body, &token)
    if err != nil {
        panic(err)
    }

    if token.Status == "success" {
        ctx.token = token.Result.Token
        ctx.UUID = token.Result.ParticipantUUID
        ctx.Expires, _ = strconv.Atoi(token.Result.Expires)
        go ctx.loginAgain()
        go ctx.subscribeEvents()
        return true
    } else {
        return false
    }
}

func (ctx * Context) RefreshToken() (bool) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/refresh_token", ctx.Conference)
    succeeded, body := ctx.Post(url, nil)
    if !succeeded {
        return false
    }
    token := GenericTokenResult{}
    if err := json.Unmarshal (body, &token); err != nil || token.Status != "success" {
        panic(err)
    }
    ctx.token = token.Result.Token
    return true
}

func (ctx * Context) SetLabel(participant Participant, label string) (succeeded bool) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/participants/%s/overlaytext", ctx.Conference, participant.UUID)
    content := fmt.Sprintf(`{"text": "%s"}`, label)
    if succeeded, _ = ctx.Post(url, []byte(content)); succeeded {
        url = fmt.Sprintf("/api/client/v2/conferences/%s/participants/%s/spotlighton", ctx.Conference, participant.UUID)
        succeeded, _ = ctx.Post(url, nil)
    }
    return succeeded
}

func (ctx * Context) Command(participant Participant, cmd string) (succeeded bool) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/participants/%s/%s", ctx.Conference, participant.UUID, cmd)
    succeeded, _ = ctx.Post(url, nil)
    return succeeded
}

func (ctx * Context) subscribeEvents() (bool) {
    url := fmt.Sprintf("/api/client/v2/conferences/%s/events", ctx.Conference)
    succeeded := ctx.listenOnEvents(url)
    if !succeeded {
        fmt.Fprintf(os.Stdout, "rest: Failed to subscribe to events\n")
    }
    return succeeded
}

