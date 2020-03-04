export function showNotification(notificationType, message) {
	const notification = document.createElement('div');
	notification.classList.add('notification', notificationType);
	notification.innerText = message;

	const notificationBox = document.getElementById('notification-box');

	const notificationNode = notificationBox.appendChild(notification);
	window.setTimeout(() => notificationBox.removeChild(notificationNode), 10000)
}