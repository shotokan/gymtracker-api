variable "name" {
  description = "Cluster name."
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version. Older versions move to paid extended support, so keep it current."
  type        = string
  default     = "1.34"
}

variable "subnet_ids" {
  description = "Subnets for the control plane and the nodes."
  type        = list(string)
}

variable "public_access_cidrs" {
  description = "CIDR blocks allowed to reach the Kubernetes API. There is no default on purpose."
  type        = list(string)

  validation {
    condition     = length(var.public_access_cidrs) > 0 && !contains(var.public_access_cidrs, "0.0.0.0/0")
    error_message = "List specific CIDR blocks. The Kubernetes API must not be open to the whole internet."
  }
}

variable "node_instance_types" {
  description = "Instance types for the node group. More than one improves Spot availability."
  type        = list(string)
  default     = ["t3.medium", "t3a.medium"]
}

variable "node_capacity_type" {
  description = "SPOT costs much less than ON_DEMAND, but AWS can reclaim the nodes."
  type        = string
  default     = "SPOT"

  validation {
    condition     = contains(["SPOT", "ON_DEMAND"], var.node_capacity_type)
    error_message = "Use SPOT or ON_DEMAND."
  }
}

variable "node_scaling" {
  description = "Node group size. A minimum of 0 lets the off-hours schedule stop every node."
  type = object({
    min     = number
    desired = number
    max     = number
  })
  default = {
    min     = 0
    desired = 2
    max     = 4
  }
}

variable "off_hours_schedule" {
  description = "Cron expressions, in UTC, that stop and restart the nodes. Null keeps them always on."
  type = object({
    stop  = string
    start = string
  })
  default = {
    stop  = "0 2 * * *"
    start = "0 13 * * 1-5"
  }
}

variable "log_retention_days" {
  description = "How long CloudWatch keeps control plane logs."
  type        = number
  default     = 7
}
