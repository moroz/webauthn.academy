download:
	go mod download

install.tools: download
	@echo Installing tools from tools.go
	@grep _ tools.go | awk -F'"' '{print $$2}' | xargs -tI % go install %
