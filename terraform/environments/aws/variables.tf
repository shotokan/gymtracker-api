variable "region" {
  description = "AWS region."
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name, used in resource names and cost tags."
  type        = string
  default     = "test"
}

variable "availability_zones" {
  description = "Availability zones to use."
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "public_access_cidrs" {
  description = "CIDR blocks allowed to reach the Kubernetes API, such as your own IP with /32."
  type        = list(string)
}

variable "monthly_budget_usd" {
  description = "Monthly budget that triggers the cost alert."
  type        = number
  default     = 50
}

variable "budget_alert_emails" {
  description = "Addresses that receive the cost alert."
  type        = list(string)

  validation {
    condition     = length(var.budget_alert_emails) > 0
    error_message = "A budget alert needs at least one address."
  }
}
