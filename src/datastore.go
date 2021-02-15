package main

// NotificationRepository is the interface to notification datastore
type NotificationRepository interface {
	Add(notification *Notification)
	Update(notification *Notification)
	Delete(notification *Notification)
	Get(id int) (Notification, error)
	GetAll() []Notification
}
