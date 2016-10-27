package fixazurerm

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"sync"
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
		},
		ResourcesMap: map[string]*schema.Resource{

			"fixazurerm_instance": instanceHi(),
		},

		ConfigureFunc: providerConfigure,
	}
}

type Config struct {
	AccessKey      string
	SubscriptionID string

	validateCredentialsOnce sync.Once
}

func instanceHi() *schema.Resource {

	return &schema.Resource{}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		AccessKey:      d.Get("access_key").(string),
		SubscriptionID: d.Get("subscription_id").(string),
	}

	log.Println("HELLOOOOOOOO!!!!")

	f, err := os.Create("/tmp/fixazurerm")
	defer f.Close()

	log.SetOutput(f)
	log.Println(fmt.Sprintf("Resource %+v", config))

	return nil, err
}
