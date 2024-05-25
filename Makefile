FORCE: ;

install:
	which modd || go install github.com/cortesi/modd/cmd/modd@latest
	which pnpm || npm i -g pnpm
	cd vite && pnpm install && cd ..

gen.chroma:
	hugo gen chromastyles --style "manni" > vite/src/assets/light.css
	hugo gen chromastyles --style "base16-snazzy" > vite/src/assets/dark.css

assets: FORCE
	cd vite && pnpm build
