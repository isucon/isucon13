export BRANCH := $(shell git branch --contains | fgrep '*' | cut -d ' ' -f 2)

.PHONY: build clean
build: #isucon12-qualify mitamae.tar.gz
	packer build honsen.pkr.hcl

app-build: 
	make -C ../../webapp/go build
# blackauth:
# 	cd ../blackauth && GOOS=linux GOARCH=amd64 go build -o ../provisioning/blackauth .

# clean:
# 	rm -rf mitamae.tar.gz isucon12-qualify* blackauth initial_data.tar.gz

# initial_data.tar.gz:
# 	gh release list | awk '/Latest/{print $$3}' | xargs gh release download --dir .

# isucon12-qualify:
# 	gh repo clone isucon/isucon12-qualify -- --depth=1 --branch=$(BRANCH)
# 	cd isucon12-qualify/bench && GOOS=linux GOARCH=amd64 make

# isucon12-qualify.tar.gz: isucon12-qualify
# 	cd isucon12-qualify && tar czf ../isucon12-qualify.tar.gz .

# mitamae.tar.gz: initial_data.tar.gz blackauth
# 	install blackauth mitamae/cookbooks/blackauth/
# 	tar czvf mitamae.tar.gz mitamae/
