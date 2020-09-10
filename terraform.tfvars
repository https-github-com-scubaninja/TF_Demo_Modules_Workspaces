application = "VMSSDemo"
environment = "modules"
location    = "westeurope"
capacity    = 5

default_tags = {
  application = "VMSSDemo"
  environment = "modules"
  deployed_by = "terraform"
}

address_space = "10.134.0.0/16"
subnet        = "10.134.20.0/24"
