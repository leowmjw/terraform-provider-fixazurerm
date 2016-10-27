default: build plan

deps:
	go install github.com/hashicorp/terraform

build:
	go build -o terraform-provider-fixazurerm .

test:
	TF_ACC=1 go test -v ./fixazurerm

plan:
	@terraform plan
