# least privilege: consumer อ่านและลบ message จาก queue นี้ได้ ไม่มีสิทธิ์อื่น
data "aws_iam_policy_document" "consumer" {
  statement {
    sid    = "ReadClaimsQueue"
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:ChangeMessageVisibility",
    ]
    resources = [module.claims_queue.arn]
  }
}

# producer ส่งเข้า queue ได้อย่างเดียว และเขียน/อ่านเฉพาะ bucket นี้
data "aws_iam_policy_document" "producer" {
  statement {
    sid       = "SendToClaimsQueue"
    effect    = "Allow"
    actions   = ["sqs:SendMessage"]
    resources = [module.claims_queue.arn]
  }

  statement {
    sid       = "ReadWriteDocs"
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["${aws_s3_bucket.docs.arn}/*"]
  }
}

# role ที่ ECS task จะ assume (เทียบได้กับ IRSA บน EKS)
data "aws_iam_policy_document" "ecs_tasks_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "consumer" {
  name   = "${var.name_prefix}-claims-consumer"
  policy = data.aws_iam_policy_document.consumer.json
  tags   = local.tags
}

resource "aws_iam_policy" "producer" {
  name   = "${var.name_prefix}-claims-producer"
  policy = data.aws_iam_policy_document.producer.json
  tags   = local.tags
}

resource "aws_iam_role" "consumer" {
  name               = "${var.name_prefix}-claims-consumer"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume.json
  tags               = local.tags
}

resource "aws_iam_role_policy_attachment" "consumer" {
  role       = aws_iam_role.consumer.name
  policy_arn = aws_iam_policy.consumer.arn
}
