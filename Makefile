refresh: ## Update Composer dependencies
	cp -r client dist/

build:
	cp -r client dist/
	cp -r webassets dist/
	go build -o ./dist/podgrab ./main.go
	./dist/podgrab

