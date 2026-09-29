package probe

import "log/slog"

// Created logs a created user with a static message and structured fields.
func Created(userID string) {
	slog.Info("user created", slog.String("user_id", userID))
	slog.Info("user created", "user_id", userID)
}
