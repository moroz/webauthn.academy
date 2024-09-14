FORCE: ;

install:
	which modd || go install github.com/cortesi/modd/cmd/modd@latest
	which pnpm || npm i -g pnpm
	cd vite && pnpm install && cd ..

assets: FORCE
	cd vite && pnpm build

highlight:
	deno run --allow-read --allow-write --allow-env vite/highlight-files.mjs

test: assets build

build:
	hugo
	make highlight
