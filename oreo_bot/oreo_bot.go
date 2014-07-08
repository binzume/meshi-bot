package main

import (
    "os"
    "log"
    "strings"
    "encoding/json"
    "io/ioutil"
    // go get github.com/thoj/go-ircevent
    "github.com/thoj/go-ircevent"
    // go get code.google.com/p/go.text/encoding
    // go get code.google.com/p/go.text/encoding/japanese
    "code.google.com/p/go.text/encoding/japanese"
    "code.google.com/p/go.text/transform"
)

var settings struct {
    Server string `json:"server"`
    UserName  string `json:"user"`
    NickName  string `json:"nick"`
    Channels  []string `json:"channels"`
    JisEncode bool `json:"jis_encode"`
}

func main() {
    configFile, _ := os.Open("settings.json")
    jsonParser := json.NewDecoder(configFile)

    if err := jsonParser.Decode(&settings); err != nil {
        log.Println("failed: load config.", err.Error())
    }
    configFile.Close()

    con := irc.IRC(settings.NickName, settings.UserName)
    err := con.Connect(settings.Server)
    if err != nil {
        log.Println("failed: connect irc")
        return
    }

    con.AddCallback("001", func (e *irc.Event) {
        for _, ch := range settings.Channels {
            con.Join(ch)
        }
    })

    // auto join
    con.AddCallback("INVITE", func (e *irc.Event) {
        //log.Println("invited " + e.Arguments[0] + "," + e.Arguments[1])
        con.Join(e.Arguments[1])
    })

    con.AddCallback("KICK", func (e *irc.Event) {
        if e.Arguments[1] == settings.NickName {
            log.Println("KICKED " + strings.Join(e.Arguments, " "))
        }
    })

    con.AddCallback("PRIVMSG", func (e *irc.Event) {
        log.Println("PRIVMSG " + strings.Join(e.Arguments, " "))
        ch := e.Arguments[0]
        msgString := string(toUtf8(e.Message()))
        log.Println("MSG " + e.Nick + ":" + msgString)

        if strings.Contains(msgString, "oreo") {
            con.Notice(ch, fromUtf8("オレオ!"))
        }
    })

    log.Println("Start bot.")
    con.Loop()
}

func toUtf8(s string) string {
    if !settings.JisEncode {
        return s
    }
    if msg, err := ioutil.ReadAll(
        transform.NewReader(strings.NewReader(s), japanese.ISO2022JP.NewDecoder())); err == nil {
        return string(msg)
    }
    return ""
}

func fromUtf8(s string) string {
    if !settings.JisEncode {
        return s
    }
    if msg, err := ioutil.ReadAll(
        transform.NewReader(strings.NewReader(s), japanese.ISO2022JP.NewEncoder())); err == nil {
        return string(msg)
    }
    return ""
}
