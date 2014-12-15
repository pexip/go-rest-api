package pexrest

import (
	"fmt"
    "text/tabwriter"
	"encoding/json"
	"strings"
    "bytes"
)

type RxResult struct {
	SSRC int `json:"ssrc"`
	PacketsReceived int `json:"packets-received"`
	OctetsReceived int `json:"octets-received"`
	Bitrate int `json:"bitrate"`
	RbPacketsLost int `json:"rb-packetslost"`
}

type TxResult struct {
	SSRC int `json:"ssrc"`
	PacketsSent int `json:"packets-sent"`
	PacketsLost int `json:"packets-lost"`
	OctetsSent int `json:"octets-sent"`
	Bitrate int `json:"bitrate"`
}

type FecIncoming struct {
	RedReceived int `json:"red-received"`
	OuterRecovered int `json:"outer-recovered"`
	InnerUnrecovered int `json:"inner-unrecovered"`
	InnerGaps bool `json:"inner-gaps"`
	InnerRecovered int `json:"inner-recovered"`
	Mode string `json:"mode"`
}

type IncomingResult struct {
	JitterbufferLatency int `json:"jitterbuffer-latency"`
	Codec string `json:"codec"`
	Fec FecIncoming `json:"fec"`
}

type FecResult struct {
	InnerProtected int `json:"inner-protected"`
	RedSent int `json:"red-sent"`
	Mode string `json:"mode"`
	OuterProtected int `json:"outer-protected"`
}

type OutgoingResult struct {
	Codec string `json:"codec"`
	Bitrate int `json:"bitrate"`
	Pacing int `json:"pacing"`
	Fec FecResult `json:"fec"`
}

type StatisticsResult struct {
	Status string `json:"status"`
	Result struct {
		Audio struct {
			Type string `json:"type"`
			Outgoing OutgoingResult `json:"outgoing"`
			Incoming IncomingResult `json:"incoming"`
			Tx TxResult `json:"tx"`
			Rx RxResult `json:"rx"`
			EncoderUsage string `json:"encoder-usage"`
		} `json:"0"`
		Video struct {
			Type string `json:"type"`
			Outgoing OutgoingResult `json:"outgoing"`
			Incoming IncomingResult `json:"incoming"`
			Tx TxResult `json:"tx"`
			Rx RxResult `json:"rx"`
			EncoderUsage string `json:"encoder-usage"`
		} `json:"1"`	
		Presentation struct {
			Type string `json:"type"`
			Outgoing OutgoingResult `json:"outgoing"`
			Incoming IncomingResult `json:"incoming"`
			Tx TxResult `json:"tx"`
			Rx RxResult `json:"rx"`
			EncoderUsage string `json:"encoder-usage"`
		} `json:"2"`	
	} `json:"result"`
}

func max(a int, b int) (int) {
	if a > b {
		return a
	} else {
		return b
	}
}

func (s StatisticsResult) String() (string) {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "\x0c")
    w := new(tabwriter.Writer)
    w.Init(&buffer, 0, 8, 1, '\t', 0)
    fmt.Fprintf(w, "| Audio\t| Video\t| Presentation\n")
    fmt.Fprintf(w, "| -----\t| -----\t| ------------\n")
    audio, _ := json.MarshalIndent(s.Result.Audio, "|", "    ")
    video, _ := json.MarshalIndent(s.Result.Video, "|", "    ")
    presentation, _ := json.MarshalIndent(s.Result.Presentation, "|", "    ")
    a := strings.Split(string(audio), "\n")
    v := strings.Split(string(video), "\n")
    p := strings.Split(string(presentation), "\n")
    total := max(max(len(a), len(v)), len(p))
    for i := 0; i < total; i++ {
    	if i < len(a) {
	    	fmt.Fprintf(w, "%s\t", a[i])
    	} else {
    		fmt.Fprintf(w, " \t")
    	}
       	if i < len(v) {
	    	fmt.Fprintf(w, "%s\t", v[i])
    	} else {
    		fmt.Fprintf(w, " \t")
    	}
    	if i < len(p) {
	    	fmt.Fprintf(w, "%s\n", p[i])
    	} else {
    		fmt.Fprintf(w, " \n")
    	}
    }
    w.Flush()
    return buffer.String()
}