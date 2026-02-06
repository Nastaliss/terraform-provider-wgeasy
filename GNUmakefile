default: build

build:
	go build -o terraform-provider-wgeasy

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Nastaliss/wgeasy/0.1.0/linux_amd64
	cp terraform-provider-wgeasy ~/.terraform.d/plugins/registry.terraform.io/Nastaliss/wgeasy/0.1.0/linux_amd64/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

.PHONY: default build install test testacc
