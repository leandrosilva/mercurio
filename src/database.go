package main

import (
	"errors"

	"gorm.io/gorm"
)

// SQLNotificationRepository is the concrete implementation of NotificationRepository for an SQL database
type SQLNotificationRepository struct {
	db *gorm.DB
}

// NewSQLNotificationRepository creates a new SQLNotificationRepository instance with an underlying GORM's database abstraction
func NewSQLNotificationRepository(db *gorm.DB) (*SQLNotificationRepository, error) {
	repository := &SQLNotificationRepository{
		db: db,
	}

	return repository, nil
}

// Add a notification to the SQL database
func (repository *SQLNotificationRepository) Add(notification *Notification) error {
	result := repository.db.Create(notification)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Update a notification in the SQL database
func (repository *SQLNotificationRepository) Update(notification *Notification) error {
	result := repository.db.Save(notification)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Delete a notification in the SQL database
func (repository *SQLNotificationRepository) Delete(destinationID string, id uint) error {
	result := repository.db.Delete(&Notification{DestinationID: destinationID}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Get a notification in the SQL database by its ID
func (repository *SQLNotificationRepository) Get(destinationID string, id uint) (Notification, error) {
	var notification Notification
	result := repository.db.Where(&Notification{ID: id, DestinationID: destinationID}).First(&notification)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Notification{}, ErrNotificationNotFound
		}
		return Notification{}, err
	}

	return notification, nil
}

// GetAll the notifications in the SQL database
func (repository *SQLNotificationRepository) GetAll(destinationID string) ([]Notification, error) {
	var notifications []Notification
	result := repository.db.Where(&Notification{DestinationID: destinationID}).Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}

// GetByStatus the notifications in the SQL database by its status (read/unread/all)
func (repository *SQLNotificationRepository) GetByStatus(destinationID string, status string) ([]Notification, error) {
	criteria := "destination_id = ?"
	if status == StatusUnreadNotifications {
		criteria += " AND read_at IS NULL"
	}
	if status == StatusReadNotifications {
		criteria += " AND read_at IS NOT NULL"
	}

	var notifications []Notification
	result := repository.db.Where(criteria, destinationID).Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}

// FilterBy all notifications in the SQL database by given criteria
func (repository *SQLNotificationRepository) FilterBy(destinationID string, criteria Notification) ([]Notification, error) {
	criteria.DestinationID = destinationID

	var notifications []Notification
	result := repository.db.Where(&criteria).Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}
