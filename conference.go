package pexrest

import (
    "fmt"
    "text/tabwriter"
    "os"
)

type ConferencesResult struct {
    Objects []struct {
        Id      string `json:"id"`
        Name    string `json:"name"`
        StartTime string `json:"start_time"`
        IsLocked    bool `json:"is_locked"`
    } `json:"objects"`
 }

func (c ConferencesResult) Print() {
    w := new(tabwriter.Writer)
    w.Init(os.Stdout, 0, 8, 1, '\t', 0)
    fmt.Fprintf(w, "Id\tName\tStartTime\tIs Locked?\t\n")
    for _, conference := range c.Objects {
        fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", conference.Id, conference.Name, conference.StartTime, conference.IsLocked)
    }
    w.Flush() 
}