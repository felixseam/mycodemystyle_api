package model

import "time"

type MemberContext struct {
	ID           int
	MemberUUID   string
	UserName     string
	LoginSession string
	StatusID     int16
	Exp          time.Time
	UserAgent    string
	IP           string
}

type UserContext struct {
	ID           int
	UserUUID     string
	UserName     string
	RoleID       uint64
	LoginSession string
	StatusID     int16
	Exp          time.Time
	UserAgent    string
	IP           string
}
