# Developer setup

Programs such as: https://github.com/mitranim/gow can be used to auto rebuild & restart the grendel binary whenever you save a file.

Example command: `gow -e go,gohtml run . serve frontend -c grendel.toml`

TailwindCSS is used for client side HTML styling, changing the tailwind classes will require recompiling the `frontend/public/tailwind.css` file. You can install the standalone Tailwind CLI, or use the node package, to automatically rebuild the css file. See https://tailwindcss.com/docs/installation for more details.

Example command: `npx tailwind -i frontend/base.css -o frontend/public/tailwind.css --watch -c frontend/tailwind.config.js`

## VSCode

This project uses Prettier and Tailwind VSCode plugins. Prettier uses the prettier-plugin-go-template and prettier-plugin-tailwindcss plugins.

> Note: This requires nodejs to be installed locally on your system

### Prettier:

> Due to what I believe is a bug with using global modules: the .prettierrc file requires absolute paths for the plugins. If your npm root path is not `/usr/lib/node_modules`, you will need to update the plugin paths in `frontend/.prettierrc`

1. Install the VSCode plugin: https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode
2. Run `sudo npm i -g prettier prettier-plugin-go-template prettier-plugin-tailwindcss`
3. Locate your root install path `npm root -g`
4. In VSCode, change: `prettier.prettierPath` to your root install path
    - You need to specify the index file: blah/node_modules/prettier/**index.cjs**
5. In VSCode, change: `editor.defaultFormatter` to `esbenp.prettier-vscode`
6. In VSCode, add: `files.associations` -> `*.gohtml: html`
7. Enable `editor.formatOnSave`

### Tailwind

1. Install the VSCode plugin: https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss
2. Run `sudo npm i -g tailwindcss`
3. In VSCode, add: `tailwindCSS.includeLanguages` -> `gohtml: html`
4. In VSCode, change `editor.quickSuggestions` -> `strings: on`
