# Test environment on AWS: VPC, EKS, ECR and DynamoDB.
#
# THIS ENVIRONMENT HAS NEVER BEEN APPLIED. CI only validates it and runs it against
# mocked providers. Applying it creates real resources that cost money.
#
# The state stays local until you add a remote backend. For shared use, add an s3
# backend with encryption and locking before the first apply.

terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.71"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.region

  # FinOps: every resource can be traced to a project and environment in Cost Explorer.
  default_tags {
    tags = {
      Project     = "gymtracker"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

locals {
  name = "gymtracker-${var.environment}"
}

module "network" {
  source = "../../modules/network"

  name               = local.name
  availability_zones = var.availability_zones
}

module "eks" {
  source = "../../modules/eks"

  name                = local.name
  subnet_ids          = module.network.subnet_ids
  public_access_cidrs = var.public_access_cidrs
}

module "ecr" {
  source = "../../modules/ecr"

  name = "gymtracker-api"
}

module "dynamodb" {
  source = "../../modules/dynamodb"

  table_name = local.name
}

# ── Identity of the API pods ─────────────────────────────────────────────────

data "aws_iam_policy_document" "api_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [module.eks.oidc_provider_arn]
    }

    # Only this service account can assume the role.
    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_issuer}:sub"
      values   = ["system:serviceaccount:gymtracker:gymtracker-api"]
    }

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_issuer}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "api" {
  name               = "${local.name}-api"
  assume_role_policy = data.aws_iam_policy_document.api_assume.json
}

data "aws_iam_policy_document" "api_table" {
  statement {
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = [
      module.dynamodb.table_arn,
      "${module.dynamodb.table_arn}/index/*",
    ]
  }
}

resource "aws_iam_role_policy" "api_table" {
  name   = "dynamodb"
  role   = aws_iam_role.api.id
  policy = data.aws_iam_policy_document.api_table.json
}

# ── Cost alert ───────────────────────────────────────────────────────────────

resource "aws_budgets_budget" "monthly" {
  name         = local.name
  budget_type  = "COST"
  limit_amount = tostring(var.monthly_budget_usd)
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "TagKeyValue"
    values = ["user:Project$gymtracker"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = var.budget_alert_emails
  }
}
