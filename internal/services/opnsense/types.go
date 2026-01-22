package opnsense

type UnboundHostOverride struct {
	Enabled     string `json:"enabled"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	Type        string `json:"rr"`
	Server      string `json:"server"`
	MXPriority  string `json:"mxprio"`
	MXDomain    string `json:"mx"`
	Description string `json:"description"`
	TxtData     string `json:"txtdata"`
}

type UnboundSearchHostOverrideItem struct {
	Id          string `json:"uuid"`
	Enabled     string `json:"enabled"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	Type        string `json:"rr"`
	Server      string `json:"server"`
	MXPriority  string `json:"mxprio"`
	MXDomain    string `json:"mx"`
	Description string `json:"description"`
	TxtData     string `json:"txtdata"`
}

type UnboundSearchHostOverrideRequest struct {
	Current      int    `json:"current"`
	RowCount     int    `json:"rowCount"`
	SearchPhrase string `json:"searchPhrase"`
}

type UnboundSearchHostOverrideResponse struct {
	Current  int                             `json:"current"`
	RowCount int                             `json:"rowCount"`
	Total    int                             `json:"total"`
	Rows     []UnboundSearchHostOverrideItem `json:"rows"`
}

type ServiceItem struct {
	Id          string `json:"id"`
	Locked      int    `json:"locked"`
	Running     int    `json:"running"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

type ServiceSearchResponse struct {
	Current  int           `json:"current"`
	RowCount int           `json:"rowCount"`
	Total    int           `json:"total"`
	Rows     []ServiceItem `json:"rows"`
}

type ServiceSearchRequest struct {
	Current      int    `json:"current"`
	RowCount     int    `json:"rowCount"`
	SearchPhrase string `json:"searchPhrase"`
}

type ServiceResponse struct {
	Status string `json:"status,omitempty"`
	Result string `json:"result,omitempty"`
}

type UnboundAddHostOverrideResponse struct {
	Result      string            `json:"result"`
	UUID        string            `json:"uuid"`
	Validations map[string]string `json:"validations"`
}

type UnboundDeleteHostOverrideResponse struct {
	Result string `json:"result"`
}
