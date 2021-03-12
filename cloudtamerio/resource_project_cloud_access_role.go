package cloudtamerio

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	hc "github.com/cloudtamer-io/terraform-provider-cloudtamerio/cloudtamerio/internal/ctclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectCloudAccessRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCloudAccessRoleCreate,
		ReadContext:   resourceProjectCloudAccessRoleRead,
		UpdateContext: resourceProjectCloudAccessRoleUpdate,
		DeleteContext: resourceProjectCloudAccessRoleDelete,
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"accounts": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"apply_to_all_accounts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"archived": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auto_pay": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aws_iam_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"aws_iam_permissions_boundary": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"aws_iam_policies": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"aws_iam_role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"azure_role_definitions": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"future_accounts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"long_term_access_keys": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"short_term_access_keys": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"user_groups": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"users": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"web_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceProjectCloudAccessRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.ProjectCloudAccessRoleCreate{
		ApplyToAllAccounts:  d.Get("apply_to_all_accounts").(bool),
		AwsIamPath:          d.Get("aws_iam_path").(string),
		AwsIamRoleName:      d.Get("aws_iam_role_name").(string),
		FutureAccounts:      d.Get("future_accounts").(bool),
		LongTermAccessKeys:  d.Get("long_term_access_keys").(bool),
		Name:                d.Get("name").(string),
		ProjectID:           d.Get("project_id").(int),
		ShortTermAccessKeys: d.Get("short_term_access_keys").(bool),
		WebAccess:           d.Get("web_access").(bool),
	}

	resp, err := c.POST("/v3/project-cloud-access-role", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceProjectCloudAccessRoleRead(ctx, d, m)

	return diags
}

func resourceProjectCloudAccessRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ProjectCloudAccessRoleResponse)
	err := c.GET(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	if hc.InflateObjectWithID(item.Accounts) != nil {
		data["accounts"] = hc.InflateObjectWithID(item.Accounts)
	}
	data["apply_to_all_accounts"] = item.ProjectCloudAccessRole.ApplyToAllAccounts
	data["archived"] = item.Project.Archived
	data["auto_pay"] = item.Project.AutoPay
	data["aws_iam_path"] = item.ProjectCloudAccessRole.AwsIamPath
	if hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary) != nil {
		data["aws_iam_permissions_boundary"] = hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary)
	}
	if hc.InflateObjectWithID(item.AwsIamPolicies) != nil {
		data["aws_iam_policies"] = hc.InflateObjectWithID(item.AwsIamPolicies)
	}
	data["aws_iam_role_name"] = item.ProjectCloudAccessRole.AwsIamRoleName
	if hc.InflateObjectWithID(item.AzureRoleDefinitions) != nil {
		data["azure_role_definitions"] = hc.InflateObjectWithID(item.AzureRoleDefinitions)
	}
	data["description"] = item.Project.Description
	data["future_accounts"] = item.ProjectCloudAccessRole.FutureAccounts
	data["long_term_access_keys"] = item.ProjectCloudAccessRole.LongTermAccessKeys
	data["name"] = item.ProjectCloudAccessRole.Name
	data["ou_id"] = item.Project.OuID
	data["project_id"] = item.ProjectCloudAccessRole.ProjectID
	data["short_term_access_keys"] = item.ProjectCloudAccessRole.ShortTermAccessKeys
	if hc.InflateObjectWithID(item.UserGroups) != nil {
		data["user_groups"] = hc.InflateObjectWithID(item.UserGroups)
	}
	if hc.InflateObjectWithID(item.Users) != nil {
		data["users"] = hc.InflateObjectWithID(item.Users)
	}
	data["web_access"] = item.ProjectCloudAccessRole.WebAccess

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set ProjectCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceProjectCloudAccessRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("apply_to_all_accounts",
		"future_accounts",
		"long_term_access_keys",
		"name",
		"short_term_access_keys",
		"web_access") {
		hasChanged++
		req := hc.ProjectCloudAccessRoleUpdate{
			ApplyToAllAccounts:  d.Get("apply_to_all_accounts").(bool),
			FutureAccounts:      d.Get("future_accounts").(bool),
			LongTermAccessKeys:  d.Get("long_term_access_keys").(bool),
			Name:                d.Get("name").(string),
			ShortTermAccessKeys: d.Get("short_term_access_keys").(bool),
			WebAccess:           d.Get("web_access").(bool),
		}

		err := c.PATCH(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update ProjectCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Handle associations.
	if d.HasChanges("accounts",
		"aws_iam_permissions_boundary",
		"aws_iam_policies",
		"azure_role_definitions",
		"user_groups",
		"users") {
		hasChanged++
		arrAddAccountIds, arrRemoveAccountIds, _, err := hc.AssociationChanged(d, "accounts")
		arrAddAwsIamPermissionsBoundary, arrRemoveAwsIamPermissionsBoundary, _, err := hc.AssociationChangedInt(d, "aws_iam_permissions_boundary")
		arrAddAwsIamPolicies, arrRemoveAwsIamPolicies, _, err := hc.AssociationChanged(d, "aws_iam_policies")
		arrAddAzureRoleDefinitions, arrRemoveAzureRoleDefinitions, _, err := hc.AssociationChanged(d, "azure_role_definitions")
		arrAddUserGroupIds, arrRemoveUserGroupIds, _, err := hc.AssociationChanged(d, "user_groups")
		arrAddUserIds, arrRemoveUserIds, _, err := hc.AssociationChanged(d, "users")

		if len(arrAddAccountIds) > 0 ||
			arrAddAwsIamPermissionsBoundary != nil ||
			len(arrAddAwsIamPolicies) > 0 ||
			len(arrAddAzureRoleDefinitions) > 0 ||
			len(arrAddUserGroupIds) > 0 ||
			len(arrAddUserIds) > 0 {
			_, err = c.POST(fmt.Sprintf("/v3/project-cloud-access-role/%s/association", ID), hc.ProjectCloudAccessRoleAssociationsAdd{
				AccountIds:                &arrAddAccountIds,
				AwsIamPermissionsBoundary: arrAddAwsIamPermissionsBoundary,
				AwsIamPolicies:            &arrAddAwsIamPolicies,
				AzureRoleDefinitions:      &arrAddAzureRoleDefinitions,
				UserGroupIds:              &arrAddUserGroupIds,
				UserIds:                   &arrAddUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on ProjectCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveAccountIds) > 0 ||
			arrRemoveAwsIamPermissionsBoundary != nil ||
			len(arrRemoveAwsIamPolicies) > 0 ||
			len(arrRemoveAzureRoleDefinitions) > 0 ||
			len(arrRemoveUserGroupIds) > 0 ||
			len(arrRemoveUserIds) > 0 {
			err = c.DELETE(fmt.Sprintf("/v3/project-cloud-access-role/%s/association", ID), hc.ProjectCloudAccessRoleAssociationsRemove{
				AccountIds:                &arrRemoveAccountIds,
				AwsIamPermissionsBoundary: arrRemoveAwsIamPermissionsBoundary,
				AwsIamPolicies:            &arrRemoveAwsIamPolicies,
				AzureRoleDefinitions:      &arrRemoveAzureRoleDefinitions,
				UserGroupIds:              &arrRemoveUserGroupIds,
				UserIds:                   &arrRemoveUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on ProjectCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceProjectCloudAccessRoleRead(ctx, d, m)
}

func resourceProjectCloudAccessRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
