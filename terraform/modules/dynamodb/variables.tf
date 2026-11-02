variable "table_name" {
  description = "Table name. Must differ from the CDK table, gym-tracker, to keep both deployments apart."
  type        = string

  validation {
    condition     = var.table_name != "gym-tracker"
    error_message = "gym-tracker belongs to the CDK stack. Use another name."
  }
}

variable "deletion_protection" {
  description = "Blocks deleting the table."
  type        = bool
  default     = false
}
