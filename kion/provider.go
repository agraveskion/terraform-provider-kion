// Package kion provides the Terraform provider.
package kion

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

var awsAccountCreationMux sync.Mutex

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Description: "The URL of a Kion installation. Example: https://kion.example.com.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KION_URL", nil),
			},
			"apikey": {
				Description: "The API key generated from Kion. Example: app_1_XXXXXXXXXXXX.",
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KION_APIKEY", nil),
			},
			"apipath": {
				Description: "The base path of the API.  Defaults to /api",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/api",
			},
			"skipsslvalidation": {
				Description: "If true, will skip SSL validation.",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KION_SKIPSSLVALIDATION", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"kion_aws_account":                 resourceAwsAccount(),
			"kion_gcp_account":                 resourceGcpAccount(),
			"kion_azure_account":               resourceAzureAccount(),
			"kion_aws_cloudformation_template": resourceAwsCloudformationTemplate(),
			"kion_aws_iam_policy":              resourceAwsIamPolicy(),
			"kion_azure_policy":                resourceAzurePolicy(),
			"kion_cloud_rule":                  resourceCloudRule(),
			"kion_compliance_check":            resourceComplianceCheck(),
			"kion_compliance_standard":         resourceComplianceStandard(),
			"kion_funding_source":              resourceFundingSource(),
			"kion_label":                       resourceLabel(),
			"kion_ou_cloud_access_role":        resourceOUCloudAccessRole(),
			"kion_project_cloud_access_role":   resourceProjectCloudAccessRole(),
			"kion_ou":                          resourceOU(),
			"kion_user_group":                  resourceUserGroup(),
			"kion_saml_group_association":      resourceSamlGroupAssociation(),
			"kion_project":                     resourceProject(),
			"kion_gcp_iam_role":                resourceGcpIamRole(),
			"kion_service_control_policy":      resourceServiceControlPolicy(),
			"kion_azure_arm_template":          resourceAzureArmTemplate(),
			"kion_azure_role":                  resourceAzureRole(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"kion_account":                     dataSourceAccount(),
			"kion_cached_account":              dataSourceCachedAccount(),
			"kion_aws_cloudformation_template": dataSourceAwsCloudformationTemplate(),
			"kion_aws_iam_policy":              dataSourceAwsIamPolicy(),
			"kion_azure_policy":                dataSourceAzurePolicy(),
			"kion_cloud_rule":                  dataSourceCloudRule(),
			"kion_compliance_check":            dataSourceComplianceCheck(),
			"kion_compliance_standard":         dataSourceComplianceStandard(),
			"kion_funding_source":              dataSourceFundingSource(),
			"kion_label":                       dataSourceLabel(),
			"kion_ou":                          dataSourceOU(),
			"kion_user_group":                  dataSourceUserGroup(),
			"kion_saml_group_association":      dataSourceSamlGroupAssociation(),
			"kion_project":                     dataSourceProject(),
			"kion_gcp_iam_role":                dataSourceGcpIamRole(),
			"kion_service_control_policy":      dataServiceControlPolicy(),
			"kion_azure_arm_template":          dataSourceAzureArmTemplate(),
			"kion_azure_role":                  dataSourceAzureRole(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	kionURL := d.Get("url").(string)
	kionAPIKey := d.Get("apikey").(string)
	kionAPIPath := d.Get("apipath").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var skipSSLValidation bool
	v, ok := d.GetOk("skipsslvalidation")
	if ok {
		t := v.(bool)
		skipSSLValidation = t
	}

	k := kionclient.NewClient(kionURL, kionAPIKey, kionAPIPath, skipSSLValidation)
	err := k.GET("/v3/me/cloud-access-role", nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Kion client",
			Detail:   "Unable to authenticate - " + err.Error(),
		})

		return nil, diags
	}

	return k, diags
}
