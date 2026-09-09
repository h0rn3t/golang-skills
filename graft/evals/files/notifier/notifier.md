# evals/files/notifier/notifier.go

- Notifier · interface · L15-L17 — Notifier
- EmailNotifier · struct · L20-L22 — EmailNotifier
- Send · method · L25-L27 — func (e *EmailNotifier) Send(ctx context.Context, recipient, message string) error
- NotifierService · struct · L30-L33 — NotifierService
- NotifierServiceFactory · struct · L36-L38 — NotifierServiceFactory
- NewNotifierServiceFactory · function · L41-L43 — func NewNotifierServiceFactory(allowList []string) *NotifierServiceFactory
- Create · method · L46-L48 — func (f *NotifierServiceFactory) Create(n Notifier) *NotifierService
- Send · method · L54-L59 — func (s *NotifierService) Send(ctx context.Context, recipient, message string) error
- containsString · function · L62-L80 — func containsString(xs []string, s string) bool
- smtpClient · struct · L82-L82 — smtpClient
- send · method · L84-L87 — func (c *smtpClient) send(ctx context.Context, recipient, message string) error
