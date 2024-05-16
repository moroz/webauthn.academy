gen.chroma:
	hugo gen chromastyles --style "catppuccin-latte" > vite/src/assets/light.css
	hugo gen chromastyles --style "base16-snazzy" > vite/src/assets/dark.css
