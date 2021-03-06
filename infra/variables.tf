variable "project_name" {
  default = "contbench"
}

variable "project_id" {
  default = "contbench"
}

variable "service_account_id" {
}

variable "region" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-a"
}

variable "functions" {
  type = map
}

variable "functions_runtime" {
  default = "go113"
}

variable "commits_collection" {
  default = "commits"
}

variable "jobs_topic" {
  default = "jobs"
}

variable "network_tier" {
  default = "STANDARD"
}
