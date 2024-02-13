### Burn After Reading

B.A.R is an open source project for those that want to self-host their own paste bin. It is built with GO + HTMX.

This project is very much a work in progress and is currently a weekend project for me to learn Go and HTMX. Please use with caution. PRs welcome.

#### Development

Install the `air` command:
```
go install github.com/cosmtrek/air@latest
```

Run the `air` command to launch the dev server:
```
air
```

Download the [Tailwindcss Standalone CLI](https://tailwindcss.com/blog/standalone-cli):

In a separate window, run the tailwindcss build command:
```
tailwindcss build -o public/css/tailwind.css --watch
```
