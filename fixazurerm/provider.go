package fixazurerm

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/terraform/helper/mutexkv"
	riviera "github.com/jen20/riviera/azure"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "bob!!",
			},
			"subscription_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
			},

			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},

			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{

			"fixazurerm_instance": instanceHi(),
		},

		ConfigureFunc: providerConfigure,
	}
}

type Config struct {
	AccessKey string

	ManagementURL string

	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string

	validateCredentialsOnce sync.Once
}

func instanceHi() *schema.Resource {

	return &schema.Resource{}
}

func (c *Config) validate() error {
	var err *multierror.Error

	if c.SubscriptionID == "" {
		err = multierror.Append(err, fmt.Errorf("Subscription ID must be configured for the AzureRM provider"))
	}
	if c.ClientID == "" {
		err = multierror.Append(err, fmt.Errorf("Client ID must be configured for the AzureRM provider"))
	}
	if c.ClientSecret == "" {
		err = multierror.Append(err, fmt.Errorf("Client Secret must be configured for the AzureRM provider"))
	}
	if c.TenantID == "" {
		err = multierror.Append(err, fmt.Errorf("Tenant ID must be configured for the AzureRM provider"))
	}

	return err.ErrorOrNil()
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		AccessKey:      d.Get("access_key").(string),
		SubscriptionID: d.Get("subscription_id").(string),
		ClientID:       d.Get("client_id").(string),
		ClientSecret:   d.Get("client_secret").(string),
		TenantID:       d.Get("tenant_id").(string),
	}

	log.Println("HELLOOOOOOOO!!!!")

	f, err := os.Create("/tmp/fixazurerm")
	defer f.Close()

	log.SetOutput(f)
	log.Println(fmt.Sprintf("Resource %+v", config))

	if err := config.validate(); err != nil {
		return nil, err
	}

	client, err := config.getArmClient()
	if err != nil {
		return nil, err
	}

	err = registerAzureResourceProvidersWithSubscription(client.rivieraClient)
	if err != nil {
		return nil, err
	}

	return client, nil
}

var providerRegistrationOnce sync.Once

// registerAzureResourceProvidersWithSubscription uses the providers client to register
// all Azure resource providers which the Terraform provider may require (regardless of
// whether they are actually used by the configuration or not). It was confirmed by Microsoft
// that this is the approach their own internal tools also take.
func registerAzureResourceProvidersWithSubscription(client *riviera.Client) error {
	var err error
	providerRegistrationOnce.Do(func() {
		// We register Microsoft.Compute during client initialization
		providers := []string{
			"Microsoft.Network", "Microsoft.Cdn", "Microsoft.Storage", "Microsoft.Sql",
			"Microsoft.Search", "Microsoft.Resources", "Microsoft.ServiceBus",
		}

		var wg sync.WaitGroup
		wg.Add(len(providers))
		for _, providerName := range providers {
			go func(p string) {
				defer wg.Done()
				if innerErr := registerProviderWithSubscription(p, client); err != nil {
					err = innerErr
				}
			}(providerName)
		}
		wg.Wait()
	})

	return err
}

// azureRMNormalizeLocation is a function which normalises human-readable region/location
// names (e.g. "West US") to the values used and returned by the Azure API (e.g. "westus").
// In state we track the API internal version as it is easier to go from the human form
// to the canonical form than the other way around.
func azureRMNormalizeLocation(location interface{}) string {
	input := location.(string)
	return strings.Replace(strings.ToLower(input), " ", "", -1)
}

// armMutexKV is the instance of MutexKV for ARM resources
var armMutexKV = mutexkv.NewMutexKV()
