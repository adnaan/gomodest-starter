package pkg

import (
	"fmt"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/jordan-wright/email"

	"github.com/adnaan/users"
	"github.com/matcornic/hermes/v2"
)

func sendEmailFunc(cfg Config) users.SendMailFunc {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: "Gomodest",
			Link: "https://gomodest.xyz",
			//Logo: "https://github.com/matcornic/hermes/blob/master/examples/gopher.png?raw=true",
		},
	}

	pool := newEmailPool(cfg)
	return func(mailType users.MailType, token, sendTo string, metadata map[string]interface{}) error {
		var name string
		var ok bool
		if metadata["name"] != nil {
			name, ok = metadata["name"].(string)
			if !ok {
				name = ""
			}
		}

		var emailTmpl hermes.Email
		var subject string
		host := cfg.Host
		if host == "0.0.0.0" || host == "localhost" {
			host = fmt.Sprintf("%s:%d", host, cfg.Port)
		}

		switch mailType {
		case users.Confirmation:
			subject = "Welcome to Gomodest!"
			emailTmpl = confirmation(name, fmt.Sprintf("%s://%s/confirm/%s", cfg.Scheme, host, token))
		case users.Recovery:
			subject = "Reset password on Gomodest.xyz"
			emailTmpl = recovery(name, fmt.Sprintf("%s://%s/reset/%s", cfg.Scheme, host, token))
		case users.ChangeEmail:
			subject = "Change email on Gomodest.xyz"
			emailTmpl = changeEmail(name, fmt.Sprintf("%s://%s/change/%s", cfg.Scheme, host, token))
		case users.OTP:
			subject = "Magic link to log into Gomodest.xyz"
			emailTmpl = magic(name, fmt.Sprintf("%s://%s/magic-login/%s", cfg.Scheme, host, token))
		}

		res, err := h.GenerateHTML(emailTmpl)
		if err != nil {
			return err
		}

		e := &email.Email{
			To:      []string{sendTo},
			Subject: subject,
			HTML:    []byte(res),
			Headers: textproto.MIMEHeader{},
			From:    cfg.SMTPAdminEmail,
		}

		return pool.Send(e, 20*time.Second)
	}
}

func confirmation(name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"Welcome to Gomodest! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Gomodest, please click here:",
					Button: hermes.Button{
						Text: "Confirm your account",
						Link: link,
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}
}

func changeEmail(name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"You have received this email because you have requested to change the email linked to your account",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to change the email linked to your account",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Confirm email change",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"If you did not request a email change, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func recovery(name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"You have received this email because a password reset request for Gomodest account was received.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Reset your password",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func magic(name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"You have received this email because a request for a magic login link for your Gomodest account was received.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to login:",
					Button: hermes.Button{
						Text: "Login with magic link",
						Link: link,
					},
				},
			},
			Outros: []string{
				"If you did not request a magic login link, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func newEmailPool(cfg Config) *email.Pool {

	var pool *email.Pool
	var err error

	if cfg.SMTPDebug {
		pool, err = email.NewPool(
			fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
			10, &unencryptedAuth{
				smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)},
		)

		if err != nil {
			panic(err)
		}

		return pool
	}

	pool, err = email.NewPool(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		10,
		smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost),
	)

	if err != nil {
		panic(err)
	}

	return pool
}

type unencryptedAuth struct {
	smtp.Auth
}

// Start starts the auth process for the specified SMTP server.
func (u *unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	server.TLS = true
	return u.Auth.Start(server)
}
