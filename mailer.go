package main

type Mailer interface {
	SendInsights(in []*DailyInsight, u *User) error
}
