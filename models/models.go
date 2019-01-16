package models

type AppModel struct {
	Metadata MetadataModel
}
type AppsModel struct {
	Resources []AppModel `json:"resources"`
}
type MetadataModel struct {
	Guid string `json:"guid"`
}

type RevisionsModel struct {
	Pagination `json:"pagination"`
	Resources []RevisionModel `json:"resources"`
}
type RevisionModel struct {
	Guid string
	Version int
	Droplet RevisionDroplet `json:"droplet"`
}
type RevisionDroplet struct {
	Guid string
}

type DeploymentModel struct {
	Guid string
}

type ErrorsModel struct {
	Errors []ErrorModel `json:"errors"`
}
type ErrorModel struct {
	Detail string
}

type Pagination struct {
	Next Link
}
type Link struct {
	Href string
}