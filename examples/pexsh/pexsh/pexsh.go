package pexsh

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "time"
    "strconv"
    "encoding/json"
    rest "github.com/pexip/go-rest-api"
)

type PexShell struct {
    lastP rest.ParticipantsResult
    lastC rest.ConferencesResult
    ctx * rest.Context
    loggedIn bool
    info string
}

func readCommand(prompt string) string {
    if len(prompt) > 0 {
        fmt.Fprintf(os.Stdout, prompt)
    }

    reader := bufio.NewReader(os.Stdin)
    text, err := reader.ReadString('\n')
    if err != nil { panic(err) }
    return text[:len(text)-1]
}

func NewShell(display_name string, destination string) (shell PexShell) {
    shell = PexShell{}
    shell.ctx = rest.NewContext(display_name, destination)
    shell.ctx.DisplayName = display_name
    return shell
}

func (shell * PexShell) logout() {
    if shell.loggedIn {
        if shell.ctx.ReleaseToken() {
            fmt.Fprintf(os.Stdout, "Logged out\n")
            shell.loggedIn = false
        }
    }
}

func (shell * PexShell) loginToNode() bool {
    shell.ctx.Pin = readCommand("PIN: ")
    if shell.ctx.RequestToken() {
        shell.loggedIn = true
        fmt.Fprintf(os.Stdout, "pexsh: user logged in [uuid=%s,expires=%dsecs]\n", shell.ctx.UUID, shell.ctx.Expires)
        return true
    }
    return false
}

func (shell PexShell) listConferences() {
    if !shell.loggedIn {
        fmt.Fprintf(os.Stdout, "pexsh: user not logged in\n")
        return
    }

    body := shell.ctx.GetConference()
    err := json.Unmarshal (body, &shell.lastC)
    if err != nil {
        panic(err)
    }
    shell.lastC.Print()
}

func (shell * PexShell) listAllParticipants() {
    if !shell.loggedIn {
        fmt.Fprintf(os.Stdout, "pexsh: user not logged in\n")
        return
    }

    shell.lastP = shell.ctx.GetParticipants()
    fmt.Fprintf(os.Stdout, "%s", shell.lastP)
}


func (shell * PexShell) getParticipantFromUuid(uuid string) rest.Participant {
    for _, p := range shell.lastP.Objects {
        if strings.HasPrefix(p.UUID, uuid) {
            return p
        }
    }

    return rest.Participant{}
}

func (shell * PexShell) zoomInto(uuid string) {
    if participant := shell.getParticipantFromUuid(uuid); len(participant.UUID) > 0 {
        shell.info = uuid
        fmt.Fprintf(os.Stdout, "%s", participant.Dump())
    } else {
        fmt.Fprintf(os.Stdout, "pexsh: no such participant (uuid=%s)\n", uuid)
    }
}

func (shell PexShell) dump() {
    shell.lastP = shell.ctx.GetParticipants()
    if participant := shell.getParticipantFromUuid(shell.info); len(participant.UUID) > 0 {
        fmt.Fprintf(os.Stdout, "%s", participant.Dump())
    }
}

func (shell * PexShell) setLabel(label string) {
    if !shell.loggedIn {
        fmt.Fprintf(os.Stdout, "pexsh: user not logged in\n")
        return
    }

    if len(shell.info) == 0 {
        fmt.Fprintf(os.Stdout, "pexsh: need to cd into a participant first\n")
    }

    participant := shell.getParticipantFromUuid(shell.info)
    shell.ctx.SetLabel(participant, label)
}

func (shell * PexShell) command(cmd string) {
    if !shell.loggedIn {
        fmt.Fprintf(os.Stdout, "pexsh: user not logged in\n")
        return
    }

    if len(shell.info) == 0 {
        fmt.Fprintf(os.Stdout, "pexsh: need to cd into a participant first\n")
    }

    participant := shell.getParticipantFromUuid(shell.info)
    shell.ctx.Command(participant, cmd)
}

func (shell * PexShell) monitor(args []string) {
    if !shell.loggedIn {
        fmt.Fprintf(os.Stdout, "pexsh: user not logged in\n")
        return
    }

    if len(shell.info) == 0 {
        fmt.Fprintf(os.Stdout, "pexsh: need to cd into a participant first\n")
    }

    count := 10
    if len(args) > 0 {
        count, _ = strconv.Atoi(args[0])
    }

    for i := 0; i < count; i++ {
        timer := time.NewTimer(time.Second)
        <- timer.C

        participant := shell.getParticipantFromUuid(shell.info)
        succeeded, mediastatistics := shell.ctx.GetStatistics(participant)
        if !succeeded {
            timer.Stop()
            fmt.Fprintf(os.Stdout, "pexsh: %s did not match any participant\n", shell.info)
            break
        }
        fmt.Fprintf(os.Stdout, "%s%s\n", mediastatistics, strings.Repeat("*" , i))
    }

}

func (shell * PexShell) Run() {
    if !shell.loginToNode() {
        fmt.Fprintf(os.Stdout, "pexsh: failed to log in\n")
        return
    }
    for {
        prompt := fmt.Sprintf("/%s:pexsh$ ", shell.info)
        input := readCommand(prompt)
        cmd := strings.Split(input, " ")
        switch cmd[0] {
        case "ls":
            if len(cmd) <= 1 {
                shell.listConferences()
                continue
            }
            switch cmd[1] {
            case "-c":
                shell.listConferences()
            case "-p":
                shell.listAllParticipants()
            }
        case "label":
            if len(cmd) <= 1 {
                fmt.Fprintf(os.Stdout, "pexsh: label <name>\n")
                continue
            }
            shell.setLabel(cmd[1])
        case "cd":
            if len(cmd) <= 1 {
                fmt.Fprintf(os.Stdout, "pexsh: cd <participant_uuid>\n")
                continue
            }
            shell.zoomInto(cmd[1])
        case "dump":
            shell.dump()
        case "monitor":
            shell.monitor(cmd[1:])
        case "disconnect":
            shell.command("disconnect")
        case "unmute":
            shell.command("unmute")
        case "mute":
            shell.command("mute")
        case "bye":
            shell.logout()
            return
        }
    }
}