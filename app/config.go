package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Name             string `json:"name" default:"gomodest"`
	Domain           string `json:"domain" default:"https://gomodest.xyz"`
	Port             int    `json:"port" default:"4000"`
	HealthPath       string `json:"health_path" envconfig:"health_path" default:"/healthz"`
	ReadTimeoutSecs  int    `json:"read_timeout_secs" envconfig:"read_timeout_secs" default:"5"`
	WriteTimeoutSecs int    `json:"write_timeout_secs" envconfig:"write_timeout_secs" default:"10"`
	LogLevel         string `json:"log_level" envconfig:"log_level" default:"error"`
	LogFormatJSON    bool   `json:"log_format_json" envconfig:"log_format_json" default:"false"`
	Templates        string `json:"templates" envconfig:"templates" default:"templates"`
	SessionSecret    string `json:"session_secret" envconfig:"session_secret" default:"mysessionsecret"`
	APIMasterSecret  string `json:"api_master_secret" envconfig:"api_master_secret" default:"supersecretkeyyoushouldnotcommit"`

	// datasource
	Driver     string `json:"driver" envconfig:"driver" default:"sqlite3"`
	DataSource string `json:"datasource" envconfig:"datasource" default:"file:users.db?mode=memory&cache=shared&_fk=1"`

	// smtp
	SMTPHost       string `json:"smtp_host" envconfig:"smtp_host" default:"0.0.0.0"`
	SMTPPort       int    `json:"smtp_port,omitempty" envconfig:"smtp_port" default:"1025"`
	SMTPUser       string `json:"smtp_user" envconfig:"smtp_user" default:"myuser" `
	SMTPPass       string `json:"smtp_pass,omitempty" envconfig:"smtp_pass" default:"mypass"`
	SMTPAdminEmail string `json:"smtp_admin_email" envconfig:"smtp_admin_email" default:"noreply@gomodest.xyz"`
	SMTPDebug      bool   `json:"smtp_debug" envconfig:"smtp_debug" default:"true"`

	// goth
	GoogleClientID string `json:"google_client_id" envconfig:"google_client_id"`
	GoogleSecret   string `json:"google_secret" envconfig:"google_secret"`

	// subscription
	PlansFile            string `json:"plans_file" envconfig:"plans_file" default:"plans.development.json"`
	Plans                []Plan `json:"-" envconfig:"-"`
	StripePublishableKey string `json:"stripe_publishable_key" envconfig:"stripe_publishable_key"`
	StripeSecretKey      string `json:"stripe_secret_key" envconfig:"stripe_secret_key"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret" envconfig:"stripe_webhook_secret"`
}

type Plan struct {
	PriceID string   `json:"price_id"`
	Name    string   `json:"name"`
	Price   string   `json:"price"`
	Current bool     `json:"-"`
	Details []string `json:"details"`

	StripeKey string `json:"-"`
}

func LoadConfig(configFile string, envPrefix string) (Config, error) {
	var config Config
	if err := loadEnvironment(configFile); err != nil {
		return config, err
	}

	if err := envconfig.Process(envPrefix, &config); err != nil {
		return config, err
	}

	plans, err := loadPlans(config.PlansFile)
	if err == nil {
		for i := range plans {
			plans[i].StripeKey = config.StripePublishableKey
		}
		config.Plans = plans
	} else {
		fmt.Printf("err loading plan file %v, err %v \n", config.PlansFile, err)
	}

	return config, nil
}

func loadPlans(file string) ([]Plan, error) {
	if file == "" {
		return []Plan{}, nil
	}

	var data []byte
	var err error

	data, err = base64.StdEncoding.DecodeString(file) // check if string is base64 data
	if err != nil {
		data, err = ioutil.ReadFile(file) // or is a file path
		if err != nil {
			return nil, err
		}
	}

	var plans []Plan
	err = json.Unmarshal(data, &plans)
	if err != nil {
		return nil, err
	}

	return plans, nil
}

func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Load(filename)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}
