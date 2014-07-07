package main

import (
    "os"
    "fmt"
    "log"
    "strings"
    "regexp"
    "time"
    "math/rand"
    "container/heap"
    "encoding/json"
    "io/ioutil"
    // hg clone https://code.google.com/p/go.text/
    "code.google.com/p/go.text/encoding/japanese"
    "code.google.com/p/go.text/transform"
    // go get github.com/thoj/go-ircevent
    "github.com/thoj/go-ircevent"
)

var settings struct {
    Server string `json:"server"`
    UserName  string `json:"user"`
    NickName  string `json:"nick"`
    RoomNames  []string `json:"default_rooms"`
    MeshiWords []string `json:"builtin_meshi_words"`
    KwskWords []string `json:"kwsk_words"`
    InviteWords []string `json:"invite_words"`
    EffectiveSeconds int64 `json:"effective_seconds"`
    JisEncode bool `json:"jis_encode"`
}

type MeshiEvent struct {
    User string
    Message string
    Time int64
}

type MeshiEvents []MeshiEvent

func (h MeshiEvents) Len() int {
    return len(h)
}

func (h MeshiEvents) Less(i, j int) bool {
    return h[i].Time < h[j].Time
}

func (h MeshiEvents) Swap(i, j int) {
    //fmt.Printf("swap: %v <> %v\n", h[i], h[j])
    h[i], h[j] = h[j], h[i]
}

func (h *MeshiEvents) Push(x interface{}) {
    *h = append(*h, x.(MeshiEvent))
}

func (h *MeshiEvents) Pop() interface{} {
    // if len(*m) == 0 {
    //     return nil
    // }

    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]

    return x
}


func main() {
    configFile, _ := os.Open("bot.json")
    jsonParser := json.NewDecoder(configFile)

    if err := jsonParser.Decode(&settings); err != nil {
        fmt.Println("parsing config file", err.Error())
    }
    configFile.Close()
    roomName := settings.RoomNames[0]

    // init strage
    m := &MeshiEvents{}
    heap.Init(m)

    channels := make(map[string]bool)

    con := irc.IRC(settings.NickName, settings.UserName)
    err := con.Connect(settings.Server)
    if err != nil {
        fmt.Println("Failed connecting")
        return
    }

    con.AddCallback("001", func (e *irc.Event) {
        con.Join(roomName)
    })

    con.AddCallback("JOIN", func (e *irc.Event) {
        // con.Notice(e.Arguments[0], "Hello!")
        channels[e.Arguments[0]] = true
    })

    con.AddCallback("KICK", func (e *irc.Event) {
        if e.Arguments[1] == settings.NickName {
            log.Println("KICKED " + strings.Join(e.Arguments, " "))
            delete(channels, e.Arguments[0])
        }
    })

    // auto join
    con.AddCallback("INVITE", func (e *irc.Event) {
        //log.Println("invited " + e.Arguments[0] + "," + e.Arguments[1])
        con.Join(e.Arguments[1])
    })

    var lastMeshiCount = 0
    var lastMeshiTime = time.Now()
    wait := int64(300)
    threshold := 2

    con.AddCallback("PRIVMSG", func (e *irc.Event) {
        roomName := e.Arguments[0]
        log.Println("priv " + e.Arguments[0] + "," + e.Arguments[1])
        log.Println("MSG " + e.Nick + ":" + e.Message())

        if msg := toUtf8(e.Message()); msg != "" {
            msgString := string(msg)
            log.Println("MSG " + e.Nick + ":" + msgString)

            if (matchMessage(msgString, settings.MeshiWords)) {
                heap.Push(m, MeshiEvent{e.Nick, msgString, time.Now().Unix()})
            }

            // 自分あて
            if isToMe(msgString) {
                if (matchMessage(msgString, settings.KwskWords)) {
                    users := findUsers(m, e.Nick)
                    if len(users) > 0 {
                        con.Notice(roomName, "以下の方がお腹が減っていそうです " + strings.Join(users, " "))
                    } else {
                        con.Notice(roomName, "候補無し")
                    }
                } else if (matchMessage(msgString, settings.InviteWords)) {
                    users := findUsers(m, e.Nick)
                    if len(users) > 0 {
                        target := users[rand.Intn(len(users))]
                        con.Notice(roomName, "OK " + target + "を誘いました")
                        con.Privmsg(target , target + ": " + e.Nick + "がご飯に行きたがっています")
                    } else {
                        con.Notice(roomName, "候補無し")
                    }
                } else if (strings.Contains(msgString, "QUIT")) {
                    con.Quit()
                } else {
                    con.Notice(roomName, "？")
                }
            }
        }

        // 飯の機運をチェック
        count := check(m)
        if count >= threshold && (count > lastMeshiCount*2 || time.Now().Unix() > lastMeshiTime.Unix() + wait ) {
            for ch := range channels {
                con.Notice(ch, fromUtf8("飯の機運が高まっています [count:" + fmt.Sprint(count) + "]"))
            }
            lastMeshiCount = count
            lastMeshiTime = time.Now()
            // TODO: Update last notice time
        }
    })

    log.Println("Start bot.")
    con.Loop()
}


func check(h *MeshiEvents) int {
    for  h.Len() > 0 && (*h)[0].Time + settings.EffectiveSeconds < time.Now().Unix() {
        log.Println("remove! " + heap.Pop(h).(MeshiEvent).Message)
    }
    return h.Len()
}

func findUsers(h *MeshiEvents, ignore string) []string {
    u := make(map[string]bool)
    for e := range *h {
        u[(*h)[e].User] = true
    }
    var keys []string
    for k := range u {
        if k != ignore {
            keys = append(keys, k)
        }
    }
    return keys
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

func isToMe(message string) bool {
    return strings.Contains(message, settings.NickName)
}


func matchMessage(message string, words []string) bool {
    for _,str := range words {
        re, _ := regexp.Compile(str)
        log.Println(str + " " + fmt.Sprint(re.MatchString(message)) )
        if (re.MatchString(message)) {
            return true
        }
    }
    return false
}

