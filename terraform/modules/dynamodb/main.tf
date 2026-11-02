# Same schema as infra/lib/constructs/database.ts and cmd/dbinit.

terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.71"
    }
  }
}

resource "aws_dynamodb_table" "this" {
  name = var.table_name
  # FinOps: on-demand costs nothing while there is no traffic.
  billing_mode                = "PAY_PER_REQUEST"
  hash_key                    = "PK"
  range_key                   = "SK"
  deletion_protection_enabled = var.deletion_protection

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "GSI1PK"
    type = "S"
  }

  attribute {
    name = "GSI1SK"
    type = "S"
  }

  global_secondary_index {
    name            = "GSI1"
    projection_type = "ALL"

    key_schema {
      attribute_name = "GSI1PK"
      key_type       = "HASH"
    }

    key_schema {
      attribute_name = "GSI1SK"
      key_type       = "RANGE"
    }
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }
}
