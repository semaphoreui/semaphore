package server

type NotificationService interface {
	SendTestProjectNotification(projectID int) error
}

type notificationServiceImpl struct {
}

func NewNotificationService() NotificationService {
	return &notificationServiceImpl{}
}

func (n *notificationServiceImpl) SendTestProjectNotification(projectID int) (err error) {
	
	return
}
