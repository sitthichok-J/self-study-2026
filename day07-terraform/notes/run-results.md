# Day 7 run results

Run at: 2026-10-30 09:12:18  Machine: LAPTOP-MICOTO

## Summary

| Step | Result |
|---|---|
| Docker running | PASS |
| LocalStack up and healthy | PASS |
| terraform fmt -check (infra) | PASS |
| terraform init (infra) | PASS |
| terraform validate (infra) | PASS |
| terraform plan (infra) | PASS |
| terraform apply (infra) | PASS |
| plan after apply shows no changes (exit 0) | PASS |
| AWS CLI: bucket has public access fully blocked | PASS |
| AWS CLI: bucket versioning enabled | PASS |
| AWS CLI: queue has redrive policy to DLQ (maxReceiveCount 3) | PASS |
| SQS: send and receive a message | PASS |
| IAM: consumer and producer policies exist | PASS |
| terraform destroy (infra) | PASS |
| nothing left in LocalStack after destroy | PASS |
| terraform init (infra-vpc) | PASS |
| terraform validate (infra-vpc) | PASS |
| terraform plan (infra-vpc, offline, nothing applied) | PASS |

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## LocalStack up and healthy  [PASS]

```
... (746 earlier lines omitted)
 Network day07_default Creating 
 Network day07_default Creating 
 Network day07_default Created 
 Network day07_default Created 
 Container day07-localstack-1 Creating 
 Container day07-localstack-1 Created 
 Container day07-localstack-1 Starting 
 Container day07-localstack-1 Started
```

## terraform fmt -check (infra)  [PASS]

```
exit 0
... (1 earlier lines omitted)
1.9: Pulling from hashicorp/terraform
458b43bffbc3: Pulling fs layer
c3054bc0ebc2: Pulling fs layer
f7e0dccd1bf8: Pulling fs layer
5bafb31cae1b: Pulling fs layer
5bafb31cae1b: Download complete
5bafb31cae1b: Pull complete
458b43bffbc3: Download complete
f7e0dccd1bf8: Download complete
458b43bffbc3: Pull complete
f7e0dccd1bf8: Pull complete
c3054bc0ebc2: Download complete
c3054bc0ebc2: Pull complete
Digest: sha256:18f9986038bbaf02cf49db9c09261c778161c51dcc7fb7e355ae8938459428cd
Status: Downloaded newer image for hashicorp/terraform:1.9
```

## terraform init (infra)  [PASS]

```
... (6 earlier lines omitted)
- Installed hashicorp/aws v5.100.0 (signed by HashiCorp)
Terraform has created a lock file .terraform.lock.hcl to record the provider
selections it made above. Include this file in your version control repository
so that Terraform can guarantee to make the same selections by default when
you run "terraform init" in the future.

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

## terraform validate (infra)  [PASS]

```
Success! The configuration is valid.
```

## terraform plan (infra)  [PASS]

```
... (303 earlier lines omitted)
Changes to Outputs:
  + bucket_name       = "selfstudy-docs-dev01"
  + consumer_role_arn = (known after apply)
  + dlq_arn           = (known after apply)
  + queue_url         = (known after apply)

─────────────────────────────────────────────────────────────────────────────

Saved the plan to: tfplan

To perform exactly these actions, run the following command to apply:
    terraform apply "tfplan"
```

## terraform apply (infra)  [PASS]

```
... (25 earlier lines omitted)
aws_iam_policy.producer: Creation complete after 0s [id=arn:aws:iam::000000000000:policy/selfstudy-claims-producer]
aws_iam_role_policy_attachment.consumer: Creating...
aws_iam_role_policy_attachment.consumer: Creation complete after 0s [id=selfstudy-claims-consumer-20261030021558631400000001]

Apply complete! Resources: 10 added, 0 changed, 0 destroyed.

Outputs:

bucket_name = "selfstudy-docs-dev01"
consumer_role_arn = "arn:aws:iam::000000000000:role/selfstudy-claims-consumer"
dlq_arn = "arn:aws:sqs:ap-southeast-1:000000000000:selfstudy-claims-dlq"
queue_url = "http://sqs.ap-southeast-1.localhost.localstack.cloud:4566/000000000000/selfstudy-claims"
```

## plan after apply shows no changes (exit 0)  [PASS]

```
exit 0
... (13 earlier lines omitted)
aws_s3_bucket_server_side_encryption_configuration.docs: Refreshing state... [id=selfstudy-docs-dev01]
data.aws_iam_policy_document.producer: Read complete after 0s [id=2148237261]
aws_iam_policy.producer: Refreshing state... [id=arn:aws:iam::000000000000:policy/selfstudy-claims-producer]

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
```

## AWS CLI: bucket has public access fully blocked  [PASS]

```
Unable to find image 'amazon/aws-cli:latest' locally
latest: Pulling from amazon/aws-cli
2f59ed1bba09: Pulling fs layer
b233855bbfe1: Pulling fs layer
afdbc91c078a: Pulling fs layer
32073eb4d38f: Pulling fs layer
d79aba3514d5: Pulling fs layer
b233855bbfe1: Download complete
2f59ed1bba09: Download complete
afdbc91c078a: Download complete
d79aba3514d5: Download complete
32073eb4d38f: Download complete
d79aba3514d5: Pull complete
afdbc91c078a: Pull complete
32073eb4d38f: Pull complete
b233855bbfe1: Pull complete
2f59ed1bba09: Pull complete
Digest: sha256:ed0f6b551cd77b6f34e240488739e4cb90b7534409d404a0c9a2be05c4690ca1
Status: Downloaded newer image for amazon/aws-cli:latest
{
    "PublicAccessBlockConfiguration": {
        "BlockPublicAcls": true,
        "IgnorePublicAcls": true,
        "BlockPublicPolicy": true,
        "RestrictPublicBuckets": true
    }
}
```

## AWS CLI: bucket versioning enabled  [PASS]

```
{
    "Status": "Enabled"
}
```

## AWS CLI: queue has redrive policy to DLQ (maxReceiveCount 3)  [PASS]

```
{
    "Attributes": {
        "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:ap-southeast-1:000000000000:selfstudy-claims-dlq\",\"maxReceiveCount\":3}",
        "VisibilityTimeout": "60"
    }
}
```

## SQS: send and receive a message  [PASS]

```
{
    "MD5OfMessageBody": "01a5ddb22812c61e06f2ded9313c050d",
    "MessageId": "79d67640-3227-4dc6-b139-946ed4bfd9bb"
}
{
    "Messages": [
        {
            "MessageId": "79d67640-3227-4dc6-b139-946ed4bfd9bb",
            "ReceiptHandle": "ZDIyN2JkY2ItODA5OS00MDkxLTgyNDgtNmU4YTU1NjE4Y2FkIGFybjphd3M6c3FzOmFwLXNvdXRoZWFzdC0xOjAwMDAwMDAwMDAwMDpzZWxmc3R1ZHktY2xhaW1zIDc5ZDY3NjQwLTMyMjctNGRjNi1iMTM5LTk0NmVkNGJmZDliYiAxNzkzMzI2NjMyLjIxNjc5OQ==",
            "MD5OfBody": "01a5ddb22812c61e06f2ded9313c050d",
            "Body": "hello-day07"
        }
    ]
}
```

## IAM: consumer and producer policies exist  [PASS]

```
{
    "Policies": [
        {
            "PolicyName": "selfstudy-claims-consumer",
            "PolicyId": "A88IX1J510IWF8NGGK1O2",
            "Arn": "arn:aws:iam::000000000000:policy/selfstudy-claims-consumer",
            "Path": "/",
            "DefaultVersionId": "v1",
            "AttachmentCount": 1,
            "CreateDate": "2026-10-30T02:15:58.583000+00:00",
            "UpdateDate": "2026-10-30T02:15:58.583000+00:00"
        },
        {
            "PolicyName": "selfstudy-claims-producer",
            "PolicyId": "AQ108F9TD2JK2K9QZBXU6",
            "Arn": "arn:aws:iam::000000000000:policy/selfstudy-claims-producer",
            "Path": "/",
            "DefaultVersionId": "v1",
            "AttachmentCount": 0,
            "CreateDate": "2026-10-30T02:15:58.590000+00:00",
            "UpdateDate": "2026-10-30T02:15:58.590000+00:00"
        }
    ]
}
```

## terraform destroy (infra)  [PASS]

```
... (323 earlier lines omitted)
aws_s3_bucket.docs: Destroying... [id=selfstudy-docs-dev01]
aws_iam_role.consumer: Destruction complete after 0s
aws_s3_bucket.docs: Destruction complete after 0s
module.claims_queue.aws_sqs_queue.this: Destruction complete after 2s
module.claims_queue.aws_sqs_queue.dlq: Destroying... [id=http://sqs.ap-southeast-1.localhost.localstack.cloud:4566/000000000000/selfstudy-claims-dlq]
module.claims_queue.aws_sqs_queue.dlq: Destruction complete after 1s

Destroy complete! Resources: 10 destroyed.
```

## nothing left in LocalStack after destroy  [PASS]

```
{
    "Buckets": [],
    "Owner": {
        "DisplayName": "webfile",
        "ID": "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a"
    },
    "Prefix": null
}

```

## terraform init (infra-vpc)  [PASS]

```
... (9 earlier lines omitted)

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

## terraform validate (infra-vpc)  [PASS]

```
Success! The configuration is valid.
```

## terraform plan (infra-vpc, offline, nothing applied)  [PASS]

```
... (425 earlier lines omitted)
      + (known after apply),
    ]
  + public_subnet_ids  = [
      + (known after apply),
      + (known after apply),
    ]
  + vpc_id             = (known after apply)

─────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't
guarantee to take exactly these actions if you run "terraform apply" now.
```


