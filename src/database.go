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
func (repository *SQLNotificationRepository) Delete(id uint) error {
	result := repository.db.Delete(&Notification{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Get a notification in the SQL database by its ID
func (repository *SQLNotificationRepository) Get(id uint) (Notification, error) {
	var notification Notification
	result := repository.db.First(&notification, id)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Notification{}, ErrNotificationNotFound
		}
		return Notification{}, err
	}

	return notification, nil
}

// GetAll the notifications in the SQL database
func (repository *SQLNotificationRepository) GetAll() ([]Notification, error) {
	var notifications []Notification
	result := repository.db.Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}

// GetByStatus the notifications in the SQL database by its status (read/unread)
func (repository *SQLNotificationRepository) GetByStatus(read bool) ([]Notification, error) {
	criteria := "read_at is null"
	if read {
		criteria = "read_at is not null"
	}

	var notifications []Notification
	result := repository.db.Where(criteria).Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}

// GetByEventID the notifications in the SQL database by its event ID
func (repository *SQLNotificationRepository) GetByEventID(eventID string) ([]Notification, error) {
	return repository.getBy(Notification{EventID: eventID})
}

// GetBySourceID the notifications in the SQL database by its source ID
func (repository *SQLNotificationRepository) GetBySourceID(sourceID string) ([]Notification, error) {
	return repository.getBy(Notification{SourceID: sourceID})
}

// GetByDestinationID the notifications in the SQL database by its source ID
func (repository *SQLNotificationRepository) GetByDestinationID(destinationID string) ([]Notification, error) {
	return repository.getBy(Notification{SourceID: destinationID})
}

// Gets a number of notifications in the SQL database by given criteria
func (repository *SQLNotificationRepository) getBy(criteria Notification) ([]Notification, error) {
	var notifications []Notification
	result := repository.db.Where(&criteria).Find(&notifications)
	if result.Error != nil {
		return []Notification{}, result.Error
	}

	return notifications, nil
}
