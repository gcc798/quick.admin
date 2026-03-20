.PHONY: help swagger swagger-fmt run build test clean

# 默认目标
help:
	@echo "可用命令:"
	@echo "  make swagger      - 生成 Swagger 文档"
	@echo "  make swagger-fmt  - 格式化 Swagger 注释"
	@echo "  make run          - 运行服务（自动生成文档）"
	@echo "  make build        - 编译项目"
	@echo "  make test         - 运行测试"
	@echo "  make clean        - 清理生成的文件"

# 生成 Swagger 文档
swagger:
	@echo "正在生成 Swagger 文档..."
	@swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
	@echo "✅ Swagger 文档已生成: docs/swagger/"
	@echo "   - swagger.json"
	@echo "   - swagger.yaml"
	@echo "   - docs.go"

# 格式化 Swagger 注释
swagger-fmt:
	@echo "正在格式化 Swagger 注释..."
	@swag fmt -g cmd/api/main.go

# 运行服务（先生成文档）
run: swagger
	@echo "正在启动服务..."
	@go run cmd/api/main.go

# 编译项目
build: swagger
	@echo "正在编译项目..."
	@go build -o bin/nai-tizi cmd/api/main.go
	@echo "✅ 编译完成: bin/nai-tizi"

# 运行测试
test:
	@echo "正在运行测试..."
	@go test -v ./...

# 清理生成的文件
clean:
	@echo "正在清理..."
	@rm -rf docs/swagger/
	@rm -rf bin/
	@echo "✅ 清理完成"

# 安装 swag 工具
install-swag:
	@echo "正在安装 swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ swag 安装完成"
