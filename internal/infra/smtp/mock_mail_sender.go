package smtp

import (
	"context"
	"errors"

	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SentEmail struct {
	To      string
	Subject string
	Body    string
	Token   string
}

type MockMailSender struct {
	sentEmails []SentEmail
	shouldFail bool
	failError  error
}

func NewMockMailSender() *MockMailSender {
	return &MockMailSender{
		sentEmails: make([]SentEmail, 0),
	}
}

func (m *MockMailSender) SendMagicLink(ctx context.Context, tokenHash value_objects.TokenHash, to string, baseURL string) error {
	if m.shouldFail {
		if m.failError != nil {
			return m.failError
		}
		return errors.New("failed to send email")
	}

	token := tokenHash.String()
	m.sentEmails = append(m.sentEmails, SentEmail{
		To:      to,
		Subject: "Your Magic Link",
		Body:    baseURL + "/magic-links/" + token,
		Token:   token,
	})

	return nil
}

func (m *MockMailSender) GetSentEmails() []SentEmail {
	return m.sentEmails
}

func (m *MockMailSender) GetLastSentEmail() *SentEmail {
	if len(m.sentEmails) == 0 {
		return nil
	}
	return &m.sentEmails[len(m.sentEmails)-1]
}

func (m *MockMailSender) GetEmailsSentTo(to string) []SentEmail {
	result := make([]SentEmail, 0, len(m.sentEmails))

	for _, mail := range m.sentEmails {
		if mail.To == to {
			result = append(result, mail)
		}
	}
	return result
}

func (m *MockMailSender) GetSentEmailsCount() int {
	return len(m.sentEmails)
}

func (m *MockMailSender) Reset() {
	m.sentEmails = make([]SentEmail, 0)
	m.failError = nil
	m.shouldFail = false
}

func (m *MockMailSender) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func (m *MockMailSender) SetFailError(err error) {
	m.failError = err
	m.shouldFail = true
}

func (m *MockMailSender) WasSentTo(to string) bool {
	return len(m.GetEmailsSentTo(to)) > 0
}

func (m *MockMailSender) GetTokenForEmail(to string) string {
	mails := m.GetEmailsSentTo(to)
	if len(mails) > 0 {
		return mails[len(mails)-1].Token
	}
	return ""
}
