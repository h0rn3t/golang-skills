package probe

import (
	"fmt"
	"log/slog"
)

// Placed logs a placed order with the id baked into the message.
func Placed(orderID int) {
	slog.Info(fmt.Sprintf("order %d placed", orderID))
}
