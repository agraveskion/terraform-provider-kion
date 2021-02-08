package ctclient

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// FlattenStringPointer -
func FlattenStringPointer(d *schema.ResourceData, key string) *string {
	if i, ok := d.GetOk(key); ok {
		v := i.(string)
		return &v
	}

	return nil
}

// FlattenStringArray -
func FlattenStringArray(items []interface{}) []string {
	arr := make([]string, 0)
	for _, item := range items {
		v := item.(string)
		// Add this because compliance_check has an array with an empty value in: regions.
		if v != "" {
			arr = append(arr, v)
		}
	}

	return arr
}

// FlattenStringArrayPointer -
func FlattenStringArrayPointer(d *schema.ResourceData, key string) *[]string {
	if i, ok := d.GetOk(key); ok {
		v := i.([]string)
		arr := make([]string, 0)
		for _, item := range v {
			v := item
			// Add this because compliance_check has an array with an empty value in: regions.
			if v != "" {
				arr = append(arr, v)
			}
		}
		return &arr
	}

	return nil
}

// FilterStringArray -
func FilterStringArray(items []string) []string {
	arr := make([]string, 0)
	for _, item := range items {
		// Added this because compliance_check has an array with an empty value in: regions.
		if item != "" {
			arr = append(arr, item)
		}
	}

	return arr
}

// FlattenIntPointer -
func FlattenIntPointer(d *schema.ResourceData, key string) *int {
	if i, ok := d.GetOk(key); ok {
		v := i.(int)
		return &v
	}

	return nil
}

// FlattenIntArray -
func FlattenIntArray(items []interface{}) []int {
	arr := make([]int, 0)
	for _, item := range items {
		arr = append(arr, item.(int))
	}

	return arr
}

// FlattenIntArrayPointer -
func FlattenIntArrayPointer(d *schema.ResourceData, key string) *[]int {
	if i, ok := d.GetOk(key); ok {
		v := i.([]int)
		arr := make([]int, 0)
		for _, item := range v {
			arr = append(arr, item)
		}
		return &arr
	}

	return nil
}

// FlattenBoolArray -
func FlattenBoolArray(items []interface{}) []bool {
	arr := make([]bool, 0)
	for _, item := range items {
		arr = append(arr, item.(bool))
	}

	return arr
}

// FlattenBoolPointer -
func FlattenBoolPointer(d *schema.ResourceData, key string) *bool {
	if i, ok := d.GetOk(key); ok {
		v := i.(bool)
		return &v
	}

	return nil
}

// FlattenGenericIDArray -
func FlattenGenericIDArray(d *schema.ResourceData, key string) []int {
	uid := d.Get(key).([]interface{})
	uids := make([]int, 0)
	for _, item := range uid {
		v, ok := item.(map[string]interface{})
		if ok {
			uids = append(uids, v["id"].(int))
		}
	}

	return uids
}

// FlattenGenericIDPointer -
func FlattenGenericIDPointer(d *schema.ResourceData, key string) *[]int {
	uid := d.Get(key).([]interface{})
	uids := make([]int, 0)
	for _, item := range uid {
		v, ok := item.(map[string]interface{})
		if ok {
			uids = append(uids, v["id"].(int))
		}
	}

	return &uids
}

// InflateObjectWithID -
func InflateObjectWithID(arr []ObjectWithID) []interface{} {
	if arr != nil {
		final := make([]interface{}, 0)

		for _, item := range arr {
			it := make(map[string]interface{})

			it["id"] = item.ID

			final = append(final, it)
		}

		return final
	}

	return make([]interface{}, 0)
}

// FieldsChanged -
func FieldsChanged(iOld interface{}, iNew interface{}, fields []string) (map[string]interface{}, string, bool) {
	mOld := iOld.(map[string]interface{})
	mNew := iNew.(map[string]interface{})

	for _, v := range fields {
		if mNew[v] != mOld[v] {
			return mNew, v, true
		}
	}

	return mNew, "", false
}

// AssociationChanged returns arrays of which values to change.
// The fields needs to be at the top level.
func AssociationChanged(d *schema.ResourceData, fieldname string) ([]int, []int, bool, error) {
	isChanged := false

	// Get the owner users
	io, in := d.GetChange(fieldname)
	ownerOld := io.([]interface{})
	oldIDs := make([]int, 0)
	for _, item := range ownerOld {
		v, ok := item.(map[string]interface{})
		if ok {
			oldIDs = append(oldIDs, v["id"].(int))
		}
	}
	ownerNew := in.([]interface{})
	newIDs := make([]int, 0)
	for _, item := range ownerNew {
		v, ok := item.(map[string]interface{})
		if ok {
			newIDs = append(newIDs, v["id"].(int))
		}
	}

	arrUserAdd, arrUserRemove, changed := determineAssociations(newIDs, oldIDs)
	if changed {
		isChanged = true
	}

	return arrUserAdd, arrUserRemove, isChanged, nil
}

// DetermineAssociations will take in a src array (source of truth/repo) and a
// destination array (cloudtamer.io application) and then return an array of
// associations to add (arrAdd) and then remove (arrRemove).
func determineAssociations(src []int, dest []int) (arrAdd []int, arrRemove []int, isChanged bool) {
	mSrc := makeMapFromArray(src)
	mDest := makeMapFromArray(dest)

	arrAdd = make([]int, 0)
	arrRemove = make([]int, 0)
	isChanged = false

	// Determine which items to add.
	for v := range mSrc {
		if _, found := mDest[v]; !found {
			arrAdd = append(arrAdd, v)
			isChanged = true
		}
	}

	// Determine which items to remove.
	for v := range mDest {
		if _, found := mSrc[v]; !found {
			arrRemove = append(arrRemove, v)
			isChanged = true
		}
	}

	return arrAdd, arrRemove, isChanged
}

func makeMapFromArray(arr []int) map[int]bool {
	m := make(map[int]bool)
	for _, v := range arr {
		m[v] = true
	}
	return m
}
