---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cloudtamerio_ou_cloud_access_role Resource - terraform-provider-cloudtamerio"
subcategory: ""
description: |-
  
---

# Resource `cloudtamerio_ou_cloud_access_role`





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **aws_iam_role_name** (String)
- **name** (String)
- **ou_id** (Number)

### Optional

- **aws_iam_path** (String)
- **aws_iam_permissions_boundary** (Number)
- **aws_iam_policies** (Block List) (see [below for nested schema](#nestedblock--aws_iam_policies))
- **id** (String) The ID of this resource.
- **last_updated** (String)
- **long_term_access_keys** (Boolean)
- **short_term_access_keys** (Boolean)
- **user_groups** (Block List) (see [below for nested schema](#nestedblock--user_groups))
- **users** (Block List) (see [below for nested schema](#nestedblock--users))
- **web_access** (Boolean)

<a id="nestedblock--aws_iam_policies"></a>
### Nested Schema for `aws_iam_policies`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--user_groups"></a>
### Nested Schema for `user_groups`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--users"></a>
### Nested Schema for `users`

Optional:

- **id** (Number) The ID of this resource.

