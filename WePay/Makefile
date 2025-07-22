.PHONY: mock
mock:
	@mockgen -source=./internal/service/transfer.go -destination=./internal/service/mocks/transfer.go -package=svcmocks
	@mockgen -source=./internal/service/user.go -destination=./internal/service/mocks/user.go -package=svcmocks
	@mockgen -source=./internal/service/wxpay_utility/wxpay_utility.go -destination=./internal/service/mocks/wxpay_utility/wxpay_utility.go -package=wxpaymocks
	@go mod tidy