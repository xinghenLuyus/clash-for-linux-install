package main

type page struct {
	key    string
	title  string
	desc   string
	footer string
	items  []string
}

type app struct {
	home            string
	selected        int
	focused         bool
	subSelected     int
	actionOutput    string
	proxyGroups     []proxyGroup
	proxyMembers    []proxyMember
	proxyMember     bool
	proxyGroupIndex int
	profiles        []profile
	width           int
	height          int
	lastWidth       int
	lastHeight      int
	message         string
	statusCache     string
	contentCache    string
	pages           []page
	modal           *modal
	toast           string
	profilesError   string
	proxyError      string
}

type modal struct {
	Title string
	Label string
	Value string
	Error string
}

type proxyGroup struct {
	Name  string
	Type  string
	Now   string
	Count string
}

type proxyMember struct {
	Marker   string
	Name     string
	Type     string
	Provider string
	Caps     string
}

type profile struct {
	ID      string
	URL     string
	Current bool
}
