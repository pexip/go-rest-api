package pexrest

type TokenResult struct {
    Token string `json: "token"`
    ParticipantUUID string `json:"participant_uuid"`
    Role string `json: "role"`
    DisplayName string `json:"display_name"`
    Expires string `json: "expires"`
}

type GenericTokenResult struct {
    Status string `json: "status"`
    Result TokenResult `json: "result"`
}

type GenericResult struct {
    Status string `json:"status"`
}