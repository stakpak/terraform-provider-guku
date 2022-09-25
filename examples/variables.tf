variable "username" {
  type = string
}
variable "password" {
  type = string
  sensitive = true
}
variable "cluster_token" {
  type = string
  sensitive = true
}
variable "cluster_ca" {
  type = string
}
variable "cluster_server" {
  type = string
}
variable "aws_region" {
  type = string
}
variable "aws_account" {
  type = string
}
