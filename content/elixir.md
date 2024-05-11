---
title: Implementing Webauthn in Elixir
---

This section is dedicated to implementing a Webauthn registration and authentication workflow using the Phoenix framework for the Elixir programming language.
Since this is the programming language and Web framework I have worked the most over the course of my career, and since I have recently worked on such an implementation, this section may well be the most complete part of the website.

## Tutorial

The following is a step-by-step tutorial for implementing a Webauthn registration and authentication workflow using Phoenix 1.7.12.

I will be developing both the application and this website on my two Linux machines, one of them running Linux Mint 21.3, and another one running Debian 12.4:

```plain
$ uname -a
Linux lm 5.15.0-91-generic #101-Ubuntu SMP Tue Nov 14 13:30:08 UTC 2023 x86_64 x86_64 x86_64 GNU/Linux
```

On Linux and Windows, Webauthn is currently supported by Firefox and Chromium (and Chromium-based browsers, such as Google Chrome and MS Edge).
On macOS, Webauthn is also supported in Safari.

I will be testing the application on both browsers on Linux, and less frequently on macOS Monterey (I refuse to use anything post-Monterey due to the unusable System Preferences app).

## Install Phoenix

Install Elixir, Erlang, and Node using [mise](https://mise.jdx.dev/):

```plain
mise local erlang@26.2.5
mise local elixir@1.16.2-otp-26
mise local node@22.1.0
mise install
```

This tutorial has been developed with Elixir 1.16.2, Erlang 26.2.5, and Node 22.1.0. If you are reading this in the future, you may need to tweak some parts of the walkthrough.

Install Hex and Rebar:

```plain
mix do local.hex --force, local.rebar --force
```

Install the Phoenix app generator:

```shell
mix archive.install hex phx_new 1.7.12
```

## Create a project

Create a new Phoenix application without Tailwind, ESBuild, LiveView, and LiveView dashboard.
The application is called "academy," because we are developing it at the Webauthn Academy.

```shell
mix phx.new --no-assets --no-dashboard --no-live academy
```

Initialize a Git repository for the newly created project. We rename the main branch to "main," because that name seems to be the least controversial at the moment.

```plain
cd academy
git init
git branch -M main
git add .
git commit -m "Initial commit"
```

I like to build my assets with [Vite](https://vitejs.dev/), and a few years ago, I even wrote how to [integrate Vite within a Phoenix application](https://moroz.dev/blog/integrating-vite-js-with-phoenix-1-6/).
However, for this tutorial we don't need an asset bundler at all. We can use a pre-bundled CSS framework, and if we end up needing any handwritten CSS, we can use native CSS nesting---all browsers that support Webauthn support CSS nesting, anyway!

## Generate authentication workflow

Generate an authentication workflow using the `mix phx.gen.auth` generator provided by the Phoenix framework:

```plain
mix phx.gen.auth --no-live Accounts User users
```
