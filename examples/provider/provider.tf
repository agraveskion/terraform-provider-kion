provider "cloudtamerio" {
  # If these are commented out, they will be loaded from
  # environment variables.
  # url = "https://cloudtamerio.example.com"
  # apikey = "key here"
}

# Create an IAM policy.
resource "cloudtamerio_aws_iam_policy" "p1" {
  name         = "sample-resource"
  description  = "Provides read only access to Amazon EC2."
  aws_iam_path = ""
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "*",
            "Resource": "*"
        }
    ]
}
EOF
}

# Output the ID of the resource created.
output "policy_id" {
  value = cloudtamerio_aws_iam_policy.p1.id
}
