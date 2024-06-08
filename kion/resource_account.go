package kion

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceAccountRead(resource string, ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// HACK: Special case when importing existing accounts
	// When importing accounts we only have an ID and we don't know whether the
	// ID is an account ID or account_cache ID. To work around this, we allow the
	// user to import using an `account_id=` or `account_cache_id=` prefix.
	// For example:
	//   terraform import kion_aws_account.test-account account_id=123
	//   terraform import kion_aws_account.test-account account_cache_id=321
	//
	// TODO: Find a better way to determine if the imported ID is an account
	// or account cache by reading the resource value

	tflog.Debug(ctx, "Reading account information", map[string]interface{}{"resource": resource, "ID": ID})

	accountLocation, locationChanged := determineAccountLocation(ID, d)

	resp, err := fetchAccountData(client, accountLocation, ID)
	if err != nil {
		return append(diags, *err)
	}

	if locationChanged {
		if err := updateLocation(d, ID, accountLocation); err != nil {
			return append(diags, *err)
		}
	}

	data := resp.ToMap(resource)
	if err := setResourceData(d, data); err != nil {
		return append(diags, *err)
	}

	if accountLocation == ProjectLocation {
		if err := setAccountLabels(d, client, ID); err != nil {
			return append(diags, *err)
		}
	}

	return diags
}

func determineAccountLocation(ID string, d *schema.ResourceData) (string, bool) {
	switch {
	case strings.HasPrefix(ID, "account_id="):
		return ProjectLocation, true
	case strings.HasPrefix(ID, "account_cache_id="):
		return CacheLocation, true
	default:
		return getKionAccountLocation(d), false
	}
}

func fetchAccountData(client *hc.Client, accountLocation, ID string) (hc.MappableResponse, *diag.Diagnostic) {
	var accountUrl string
	var resp hc.MappableResponse

	switch accountLocation {
	case CacheLocation:
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
		resp = new(hc.AccountCacheResponse)
	case ProjectLocation:
		accountUrl = fmt.Sprintf("/v3/account/%s", ID)
		resp = new(hc.AccountResponse)
	default:
		return nil, hc.CreateDiagError("Invalid account location", fmt.Errorf("unknown account location: %s", accountLocation), ID)
	}

	if err := client.GET(accountUrl, resp); err != nil {
		return nil, hc.CreateDiagError("Unable to read account", err, ID)
	}
	return resp, nil
}

func updateLocation(d *schema.ResourceData, ID, accountLocation string) *diag.Diagnostic {
	d.SetId(ID)
	if err := d.Set("location", accountLocation); err != nil {
		return hc.CreateDiagError("Unable to set location for account", err, ID)
	}
	return nil
}

func setResourceData(d *schema.ResourceData, data map[string]interface{}) *diag.Diagnostic {
	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			return hc.CreateDiagError("Unable to read and set account", err, k)
		}
	}
	return nil
}

func setAccountLabels(d *schema.ResourceData, client *hc.Client, ID string) *diag.Diagnostic {
	labelData, err := hc.ReadResourceLabels(client, "account", ID)
	if err != nil {
		return hc.CreateDiagError("Unable to read account labels", err, ID)
	}

	if err := d.Set("labels", labelData); err != nil {
		return hc.CreateDiagError("Unable to set labels for account", err, ID)
	}
	return nil
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	oldProjectId, newProjectId := getProjectIdChanges(d)
	hasChanged := handleProjectChanges(d, client, ID, oldProjectId, newProjectId)

	if hasResourceChanges(d, "email", "name", "include_linked_account_spend", "linked_role", "skip_access_checking", "start_datecode", "use_org_account_info") {
		if err := updateAccount(d, client, ID); err != nil {
			return append(diags, *err)
		}
		hasChanged = true
	}

	if getKionAccountLocation(d) == ProjectLocation && d.HasChange("labels") {
		if err := updateAccountLabels(d, client, ID); err != nil {
			return append(diags, *err)
		}
		hasChanged = true
	}

	if hasChanged {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set last_updated",
				Detail:   fmt.Sprintf("Error: %v", err),
			})
		}
		tflog.Info(ctx, fmt.Sprintf("Updated account ID: %s", ID))
	}

	return diags
}

func getProjectIdChanges(d *schema.ResourceData) (int, int) {
	oldId, newId := d.GetChange("project_id")
	return oldId.(int), newId.(int)
}

func handleProjectChanges(d *schema.ResourceData, client *hc.Client, ID string, oldProjectId, newProjectId int) bool {
	switch {
	case oldProjectId == 0 && newProjectId != 0:
		if err := handleCacheToProjectConversion(d, client, ID, newProjectId); err != nil {
			return true
		}
	case oldProjectId != 0 && newProjectId == 0:
		if err := handleProjectToCacheConversion(d, client, ID); err != nil {
			return true
		}
	case oldProjectId != newProjectId:
		if err := moveAccountToDifferentProject(d, client, ID); err != nil {
			return true
		}
	}
	return false
}

func handleCacheToProjectConversion(d *schema.ResourceData, client *hc.Client, ID string, newProjectId int) *diag.Diagnostic {
	accountCacheId, err := strconv.Atoi(ID)
	if err != nil {
		return hc.CreateDiagError("Unable to convert cached account to project account, invalid cached account id", err, ID)
	}

	newId, err := convertCacheAccountToProjectAccount(client, accountCacheId, newProjectId, d.Get("start_datecode").(string))
	if err != nil {
		return hc.CreateDiagError("Unable to convert cached account to project account", err, ID)
	}

	d.SetId(strconv.Itoa(newId))
	if err := d.Set("location", ProjectLocation); err != nil {
		return hc.CreateDiagError("Error setting location", err, ProjectLocation)
	}
	return nil
}

func handleProjectToCacheConversion(d *schema.ResourceData, client *hc.Client, ID string) *diag.Diagnostic {
	accountId, err := strconv.Atoi(ID)
	if err != nil {
		return hc.CreateDiagError("Unable to convert project account to cache account, invalid account id", err, ID)
	}

	newId, err := convertProjectAccountToCacheAccount(client, accountId)
	if err != nil {
		return hc.CreateDiagError("Unable to convert project account to cache account", err, ID)
	}

	d.SetId(strconv.Itoa(newId))
	if err := d.Set("location", CacheLocation); err != nil {
		return hc.CreateDiagError("Unable to set location", err, CacheLocation)
	}
	return nil
}

func moveAccountToDifferentProject(d *schema.ResourceData, client *hc.Client, ID string) *diag.Diagnostic {
	req := createAccountMoveRequest(d)
	resp, err := client.POST(fmt.Sprintf("/v3/account/%s/move", ID), req)
	if err != nil {
		return hc.CreateDiagError("Unable to move account to a different project", err, ID)
	}

	d.SetId(strconv.Itoa(resp.RecordID))
	return nil
}

func createAccountMoveRequest(d *schema.ResourceData) hc.AccountMove {
	req := hc.AccountMove{
		ProjectID:        d.Get("project_id").(int),
		FinancialSetting: "move",
		MoveDate:         0,
	}
	if v, exists := d.GetOk("move_project_settings"); exists {
		moveSettings := v.(*schema.Set)
		for _, item := range moveSettings.List() {
			if moveSettingsMap, ok := item.(map[string]interface{}); ok {
				req.FinancialSetting = moveSettingsMap["financials"].(string)
				if val, ok := moveSettingsMap["move_datecode"]; ok {
					req.MoveDate = val.(int)
				}
			}
		}
	}
	return req
}

func hasResourceChanges(d *schema.ResourceData, keys ...string) bool {
	for _, key := range keys {
		if d.HasChange(key) {
			return true
		}
	}
	return false
}

func updateAccount(d *schema.ResourceData, client *hc.Client, ID string) *diag.Diagnostic {
	accountLocation := getKionAccountLocation(d)
	accountUrl := fmt.Sprintf("/v3/account/%s", ID)
	req := createAccountUpdateRequest(d, accountLocation)
	if err := client.PATCH(accountUrl, req); err != nil {
		return hc.CreateDiagError("Unable to update account", err, ID)
	}
	return nil
}

func createAccountUpdateRequest(d *schema.ResourceData, accountLocation string) interface{} {
	if accountLocation == CacheLocation {
		return hc.AccountCacheUpdatable{
			Name:                      d.Get("name").(string),
			AccountEmail:              d.Get("email").(string),
			LinkedRole:                d.Get("linked_role").(string),
			IncludeLinkedAccountSpend: hc.OptionalBool(d, "include_linked_account_spend"),
			SkipAccessChecking:        hc.OptionalBool(d, "skip_access_checking"),
		}
	}
	return hc.AccountUpdatable{
		Name:                      d.Get("name").(string),
		AccountEmail:              d.Get("email").(string),
		LinkedRole:                d.Get("linked_role").(string),
		StartDatecode:             d.Get("start_datecode").(string),
		IncludeLinkedAccountSpend: hc.OptionalBool(d, "include_linked_account_spend"),
		SkipAccessChecking:        hc.OptionalBool(d, "skip_access_checking"),
		UseOrgAccountInfo:         hc.OptionalBool(d, "use_org_account_info"),
	}
}

func updateAccountLabels(d *schema.ResourceData, client *hc.Client, ID string) *diag.Diagnostic {
	err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID)
	if err != nil {
		return hc.CreateDiagError("Unable to update account labels", err, ID)
	}
	return nil
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_ = ctx

	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	accountLocation := getKionAccountLocation(d)
	accountUrl := fmt.Sprintf("/v3/account/%s", ID)
	if accountLocation == CacheLocation {
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
	}

	err := client.DELETE(accountUrl, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete account",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	d.SetId("")
	return diags
}

func convertCacheAccountToProjectAccount(client *hc.Client, accountCacheId, newProjectId int, startDatecode string) (int, error) {
	startDatecode = strings.ReplaceAll(startDatecode, "-", "")
	resp, err := client.POST(fmt.Sprintf("/v3/account-cache/%d/convert/%d?start_datecode=%s", accountCacheId, newProjectId, startDatecode), nil)
	if err != nil {
		return 0, err
	}
	return resp.RecordID, nil
}

func convertProjectAccountToCacheAccount(client *hc.Client, accountId int) (int, error) {
	respRevert := new(hc.AccountRevertResponse)
	err := client.DeleteWithResponse(fmt.Sprintf("/v3/account/revert/%d", accountId), nil, respRevert)
	if err != nil {
		return 0, err
	}
	return respRevert.RecordID, nil
}

const (
	CacheLocation   = "cache"
	ProjectLocation = "project"
)

func getKionAccountLocation(d *schema.ResourceData) string {
	if v, exists := d.GetOk("location"); exists {
		return v.(string)
	}
	if _, exists := d.GetOk("project_id"); exists {
		return ProjectLocation
	}
	return CacheLocation
}

func customDiffComputedAccountLocation(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	var diags diag.Diagnostics

	if _, exists := d.GetOk("project_id"); exists {
		if err := d.SetNew("location", ProjectLocation); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set new computed location for project",
				Detail:   fmt.Sprintf("Error setting new computed location to ProjectLocation: %v", err),
			})
		}
	} else {
		if err := d.SetNew("location", CacheLocation); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set new computed location for cache",
				Detail:   fmt.Sprintf("Error setting new computed location to CacheLocation: %v", err),
			})
		}
	}

	if len(diags) > 0 {
		var combinedErr strings.Builder
		for _, d := range diags {
			combinedErr.WriteString(d.Detail + "\n")
		}
		return fmt.Errorf(combinedErr.String())
	}
	return nil
}
