package kion

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The field name whose values you wish to filter by.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"values": {
							Description: "The values of the field name you specified.",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"regex": {
							Description: "Dictates if the values provided should be treated as regular expressions.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"list": {
				Description: "This is where Kion makes the discovered data available as a list of resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_number": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"linked_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"account_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"payer_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"start_datecode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"skip_access_checking": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"use_org_account_info": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"linked_account_number": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"include_linked_account_spend": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"car_external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.AccountListResponse)
	err := client.GET("/v3/account", resp)
	if err != nil {
		diags = append(diags, *hc.CreateDiagError(
			"Unable to read Account",
			err,
			"all",
		))
		return diags
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := map[string]interface{}{
			"created_at":                   item.CreatedAt,
			"id":                           item.ID,
			"name":                         item.Name,
			"account_number":               item.AccountNumber,
			"email":                        item.Email,
			"linked_role":                  item.LinkedRole,
			"project_id":                   item.ProjectID,
			"account_type_id":              item.AccountTypeID,
			"payer_id":                     item.PayerID,
			"start_datecode":               item.StartDatecode,
			"skip_access_checking":         item.SkipAccessChecking,
			"use_org_account_info":         item.UseOrgAccountInfo,
			"linked_account_number":        item.LinkedAccountNumber,
			"include_linked_account_spend": item.IncludeLinkedAccountSpend,
			"car_external_id":              item.CARExternalID,
			"service_external_id":          item.ServiceExternalID,
		}

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, *hc.CreateDiagError(
				"Unable to filter Account",
				err,
				"filter",
			))
			return diags
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		diags = append(diags, *hc.CreateDiagError(
			"Unable to read Account",
			err,
			"all",
		))
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
