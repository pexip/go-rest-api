package pexrest

import (
    "fmt"
    "text/tabwriter"
    "reflect"
    "bytes"
)

type Participant struct {
        UUID string `json:"uuid"`
        URI string `json:"uri"`
        DisplayName string `json:"display_name"`
        HasMedia bool `json:"has_media"`
        StartTime uint `json:"start_time"`
        Vad uint `json:"vad"`
        IsMuted string `json:"is_muted"`
        IsPresenting      string `json:"is_presenting"`
        RxPresentationPolicy string `json:"rx_presentation_policy"`
        Role string `json:"role"`
        ServiceType string `json:"service_type"`
        Type string `json:"type"`
        PresentationSupported string `json:"presentation_supported"`
        StageIndex int `json:"stage_index"`
        OverlayText string `json:"overlay_text"`
        Spotlight float64 `json:"spotlight"`

}

type ParticipantsResult struct {
    Status string `json:"statusx"`
    Objects []Participant `json:"result"`
}

func (p ParticipantsResult) String() (string) {
    w := new(tabwriter.Writer)
    var buffer bytes.Buffer
    w.Init(&buffer, 0, 8, 1, '\t', 0)
    fmt.Fprintf(w, "Id\tType\tUri\tDisplay Name\tMedia?\t\n")
    for _, participant := range p.Objects {
        fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t\n", participant.UUID, participant.Type, participant.URI, participant.DisplayName, participant.HasMedia)
    }
    w.Flush()
    return buffer.String()
}

func (p Participant) Dump() (string) {
    w := new(tabwriter.Writer)
    var buffer bytes.Buffer
    w.Init(&buffer, 0, 8, 1, '\t', 0)
    s := reflect.ValueOf(&p).Elem()
    typeOfP := s.Type()
    fmt.Fprintf(w, "| Attr\tValue\t|\tAttr\tValue\n")
    fmt.Fprintf(w, "| ----\t-----\t|\t----\t-----\n")
    for i := 0; i < s.NumField(); i+=2 {
        f := s.Field(i)
        if i + 1 < s.NumField() {
            f2 := s.Field(i + 1)
            fmt.Fprintf(w, "| %s\t%v\t|\t%s\t%s\n", typeOfP.Field(i).Name, f.Interface(),  typeOfP.Field(i + 1).Name, f2.Interface())
        } else {
            fmt.Fprintf(w, "| %s\t%v\t|\n", typeOfP.Field(i).Name, f.Interface())
        }
    }
    w.Flush()
    return buffer.String()
}
