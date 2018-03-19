package database

const (
	StatusAvailable Status = iota
	StatusUnavailable
	StatusErrored
)

const (
	CredTypeAdmin CredentialType = iota
	CredTypeAppUser
	CredTypeAppAdmin
	CredTypeAppReadOnly
	CredTypeMonitoring
)

func GetAllCredentialTypes() []CredentialType {
	return []CredentialType{CredTypeAdmin, CredTypeAppUser, CredTypeAppAdmin, CredTypeAppReadOnly, CredTypeMonitoring}
}

func GetUserNameForType(t CredentialType) string {
	switch t {
	case CredTypeAdmin:
		return "master"
	case CredTypeAppUser:
		return "appuser"
	case CredTypeAppAdmin:
		return "appadmin"
	case CredTypeMonitoring:
		return "monitoring"
	default:
		return "appreadonly"
	}
}

const (
	SizeXSmall Size = iota
	SizeSmall
	SizeMedium
	SizeLarge
	SizeXLarge
	SizeXXLarge
	SizeMassive
)

type Size int

type Status int

type Password string

type CredentialType int

type CredentialID string

type Credentials map[CredentialType]*Credential

type DatabaseID string

type Scope string

type Request struct {
	ID       DatabaseID
	Name     string
	Storage  int64
	Size     Size
	HA       bool
	Metadata map[string]string
	Owner    string
}

type StatusRequest struct {
	Name string
	Status
	ID *DatabaseID
	Scope
}

type Credential struct {
	ID           CredentialID
	Username     string
	Password     Password
	Host         string
	Port         int64
	DatabaseName string
	CredType     CredentialType
	Scope
}

type Database struct {
	ID          DatabaseID
	Storage     int64
	Size        Size
	Status      Status
	HA          bool
	Credentials Credentials
	Name        string
	Host        string
	Port        int64
	Owner       string
}

func GetMessageForStatus(s Status) string {
	switch s {
	case StatusAvailable:
		return "db is available"
	case StatusUnavailable:
		return "db currently unavailable"
	case StatusErrored:
		return "unable to create db"
	default:
		return "db currently unavailable"
	}

}
