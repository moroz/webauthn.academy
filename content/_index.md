---
title: Homepage
---

## What is Webauthn and why should I care?

Webauthn (short for _Web Authentication API_) is a technology developed by the [W3C](https://www.w3.org/) and the [FIDO Alliance](https://fidoalliance.org/).
It provides a way to securely sign in to websites using public key cryptography.
Webauthn can be used to replace a password completely (a use case commonly called "signing in with a passkey"), or as an additional security measure on top of password authentication.

As of 2024, Webauthn is supported by [all major desktop and mobile browsers](https://caniuse.com/webauthn) and across all three major desktop operating systems (GNU/Linux, MS Windows, and macOS).

## The purpose of this website

The aim of this website is to describe the process of integrating Webauthn on your website.
I will try to keep the instructions mostly backend-agnostic, and if I do present any examples of back end code, they will be written in the languages I am most familiar with (Elixir and Go).
JavaScript usage is mandatory, as the Web Authentication API in browsers is only exposed through a JavaScript interface.

When I tried to implement Webauthn in my own projects, I found out that there are no comprehensive learning resources for developers interested in this technology.
The two main websites dedicated to Webauthn, [webauthn.io](https://webauthn.io/) and [webauthn.guide](https://webauthn.guide/) present only a superficial view of Webauthn, and omit most of the technical details completely.
Even the documentation and demo projects for Webauthn in specfic programming languages tend not to explain many of the important parts of the integrations, such as:

- How to parse and serialize binary data sent by the back end?
- How to store the public keys and authenticator metadata in a database?
- What are all those cryptic options in the JavaScript APIs (`navigator.credentials.create` and `navigator.credentials.get`)?
- How do these options influence the security of my application?
- Who the heck is Alice?

Starting from May 7th, 2024, I intend to spend an hour each day working on this website, for at least a month.
