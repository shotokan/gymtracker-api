variable "name" {
  description = "Prefix for resource names."
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block of the VPC."
  type        = string
  default     = "10.40.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones for the subnets. EKS requires at least two."
  type        = list(string)

  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "EKS requires subnets in at least two availability zones."
  }
}
