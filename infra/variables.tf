variable "project_name" {
  default = "contbench"
}

variable "project_id" {
  default = "contbench"
}

variable "region" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-a"
}

variable "functions" {
  type = list(object({
    name         = string,
    trigger_type = string,
  }))
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
