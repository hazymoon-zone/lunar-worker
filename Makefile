.PHONY: test test-integration

test:
	go test $$(go list ./... | grep -v '/api$$') -skip '^TestIntegration'

test-integration:
	encore test ./... -run '^TestIntegration'
