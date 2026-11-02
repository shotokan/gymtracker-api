# Runs against mocked providers: it needs no credentials and creates nothing.

mock_provider "aws" {
  mock_data "aws_iam_policy_document" {
    defaults = {
      json = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"
    }
  }

  mock_resource "aws_iam_role" {
    defaults = {
      arn = "arn:aws:iam::123456789012:role/mock"
    }
  }

  mock_resource "aws_iam_openid_connect_provider" {
    defaults = {
      arn = "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/MOCK"
    }
  }

  mock_resource "aws_kms_key" {
    defaults = {
      arn = "arn:aws:kms:us-east-1:123456789012:key/mock"
    }
  }

  mock_resource "aws_dynamodb_table" {
    defaults = {
      arn = "arn:aws:dynamodb:us-east-1:123456789012:table/mock"
    }
  }

  mock_resource "aws_eks_cluster" {
    defaults = {
      identity = [{ oidc = [{ issuer = "https://oidc.eks.us-east-1.amazonaws.com/id/MOCK" }] }]
    }
  }

  mock_resource "aws_eks_node_group" {
    defaults = {
      resources = [{ autoscaling_groups = [{ name = "mock-asg" }], remote_access_security_group_id = "" }]
    }
  }
}

mock_provider "tls" {
  mock_data "tls_certificate" {
    defaults = {
      certificates = [{ sha1_fingerprint = "0000000000000000000000000000000000000000" }]
    }
  }
}

variables {
  public_access_cidrs = ["203.0.113.10/32"]
  budget_alert_emails = ["alerts@example.com"]
}

run "plans_every_module" {
  command = apply

  assert {
    condition     = module.dynamodb.table_name == "gymtracker-test"
    error_message = "The table must not collide with the CDK table."
  }

  assert {
    condition     = length(module.network.subnet_ids) == 2
    error_message = "EKS needs subnets in two availability zones."
  }
}

run "rejects_an_open_kubernetes_api" {
  command = plan

  module {
    source = "../../modules/eks"
  }

  variables {
    name                = "test"
    subnet_ids          = ["subnet-1", "subnet-2"]
    public_access_cidrs = ["0.0.0.0/0"]
  }

  expect_failures = [var.public_access_cidrs]
}

run "rejects_the_cdk_table_name" {
  command = plan

  module {
    source = "../../modules/dynamodb"
  }

  variables {
    table_name = "gym-tracker"
  }

  expect_failures = [var.table_name]
}
