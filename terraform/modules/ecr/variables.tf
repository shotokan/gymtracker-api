variable "name" {
  description = "Repository name."
  type        = string
}

variable "images_to_keep" {
  description = "How many tagged images to keep."
  type        = number
  default     = 10
}
