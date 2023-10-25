build-local:
	go build -o dist/meter-reader-macos *.go && chmod +x dist/meter-reader-macos

build-remote:
	GOOS=linux GOARCH=arm GOARM=5 go build -o dist/meter-reader *.go && chmod +x dist/meter-reader

deploy: build-remote kill-remote
	scp dist/meter-reader kamstrup-meter-reader:/home/erikbooij/meter-reader-server && ssh kamstrup-meter-reader sudo systemctl restart meter-reader

kill-remote:
	ssh kamstrup-meter-reader sudo pkill meter-reader || true

run-local: build-local
	./dist/meter-reader-macos --port -tbd-

run-remote: build-remote upload-tmp
	ssh -t kamstrup-meter-reader sudo /tmp/meter-reader --port /dev/ttyUSB0

upload-tmp:
	scp dist/meter-reader kamstrup-meter-reader:/tmp/meter-reader

run-remote-log:
	make run-remote > comparison-output.txt

compare:
	md5 -q base-output.txt && md5 -q comparison-output.txt
