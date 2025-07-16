![](_docs/img/logo.jpeg)

**Dreaming Bard** is your assistant to create long books/stories/documents.

The goal is to **research** about long LLM context and test solutions which can help maintain conversations.

The ultimate goal - to be able to write a standard-length novel book: around 65k-80k words. And do it without a hole in
your pocket.

It has some number of interesting features:

- Has chats where you can attach page(s)/summary(ies) or lore to discuss
- Any message can be used stored as a new page or as a new lore card
- You can define multiple **"Roles"**: LLM cards with specific model and system prompt
- You can save (and pin) pre-filled chat templates: they are called "Prompts".
- You can plan the next page using **"Blueprints"**: a combination of page outline, linked context and pages to generate
  a new page.
- A final book can be exported as ePub
- Almost everything can be imported or exported as plain Markdown files

Tech features

- Lightweight: memory footprint around 20MiB
- Simple as hell: one binary and one state file (SQLite DB).
- Cross-Platform: pre-built binaries for ARM and AMD64
- Supports OIDC (optional)
- Supports OpenAI (or compatible) API, Gemini (AI Studio), Ollama
- Can work without JavaScript (though few functions will be not available)
- Mobile friendly design (at least for my phone, and at least friendly to me)

**Limitations**

- Focused on a chunked text sequence (pages). Not a general-purpose solution like alternatives, by-design (at least for
  now)
- (temporary) only a single book (collection of pages) supported
- (temporary) only a single provider allowed (but no limits on models)
- Single tenant (by-design)
- Requires external LLM provider (by-design)
- No embedding database (by-design): embedding search doesn't work well for sequenced text like pages

**Alternatives**

Writing/RP

- [SillyTavern](https://github.com/SillyTavern/SillyTavern)

All In One

- [Openweb UI](https://openwebui.com/)
- [Librechat](https://www.librechat.ai/)

## Status

This is a research project; UI and backend quality are definitely below my own standards, but everything should
work.

Mostly, updates should be safe but only forward. Please, do back up your database before every upgrade (however, patch
updates should be relatively safe)

Feel free to [contribute](CONTRIBUTING.md)!

## Reasoning

I like books. Like really like to read them.

Once upon a time I realized that there are not enough books of type (not __that__ type) I want to read.

It started as a research project to handle long (hundreds pages) book writing. It's intentionally not using any kind
of AI/LLM frameworks to understand better how LLM works and how it handles context.

## Quick start

(not the only way, but most trivial)

- Get [docker](https://docs.docker.com/get-started/get-docker/) 
- Get API key from [https://openai.com/]
- Run



    docker run -v "$(pwd):/data" \
      --rm -p 8080:8080 \
      -e PROVIDER_TYPE=openai \
      -e PROVIDER_OPENAI_TOKEN=sk-SUPER-SECRET-TOKEN \
      ghcr.io/reddec/dreaming-bard:latest


- Open http://localhost:8080


## Installation

- [Binary releases](https://github.com/reddec/dreaming-bard/releases/latest) for all major platforms
- [Docker](https://github.com/reddec/dreaming-bard/pkgs/container/dreaming-bard) containers (arm64 and amd64)
- [Docker Compose](docker-compose.yaml) (arm64 and amd64)
- From source `go install github.com/reddec/dreaming-bard@latest`

## Run

Web interface available via http://localhost:8080

### Docker

    docker run --rm -v "$(pwd):/data" -p 8080:8080 ghcr.io/reddec/dreaming-bard:latest

### Docker Compose

- Download [docker-compose.yaml](docker-compose.yaml)
- In the directory: `docker compose up`

### CLI

(download from releases or build it by yourself)

    dreaming-bard server

<details>
<summary>See full CLI reference</summary>

```
Usage: dreaming-bard server [flags]

Run server

Flags:
  -h, --help                                                  Show context-sensitive help.
  -C, --change-dir=STRING                                     Change directory ($CHANGE_DIR)

      --cors                                                  Enable CORS ($CORS)
      --bind=":8080"                                          Binding address ($BIND)
      --disable-gzip                                          Disable gzip compression for HTTP ($DISABLE_GZIP)
      --parallel-workers=1                                    Number of parallel workers (chats) ($PARALLEL_WORKERS)
      --provider-type="ollama"                                Provider type ($PROVIDER_TYPE)
      --provider-openai-url="https://api.openai.com/v1"       OpenAI base URL ($PROVIDER_OPENAI_URL)
      --provider-openai-model="gpt-4o"                        OpenAI model name ($PROVIDER_OPENAI_MODEL)
      --provider-openai-token=STRING                          OpenAI API token ($PROVIDER_OPENAI_TOKEN)
      --provider-openai-timeout=3m                            Timeout ($PROVIDER_OPENAI_TIMEOUT)
      --provider-openai-max-tokens=8192                       Max tokens ($PROVIDER_OPENAI_MAX_TOKENS)
      --provider-openai-temperature=0.8                       Temperature ($PROVIDER_OPENAI_TEMPERATURE)
      --provider-openai-top-p=0.9                             Top P ($PROVIDER_OPENAI_TOP_P)
      --provider-ollama-url="http://localhost:11434"          Ollama OpenAPI URL ($PROVIDER_OLLAMA_URL)
      --provider-ollama-model="qwen3:14b"                     Ollama model name ($PROVIDER_OLLAMA_MODEL)
      --provider-ollama-timeout=120s                          Timeout ($PROVIDER_OLLAMA_TIMEOUT)
      --provider-ollama-context-size=32768                    Context size ($PROVIDER_OLLAMA_CONTEXT_SIZE)
      --provider-ollama-max-tokens=32768                      Max tokens ($PROVIDER_OLLAMA_MAX_TOKENS)
      --provider-ollama-temperature=0.6                       Temperature ($PROVIDER_OLLAMA_TEMPERATURE)
      --provider-ollama-top-p=0.95                            Top P ($PROVIDER_OLLAMA_TOP_P)
      --provider-ollama-top-k=20                              Top K ($PROVIDER_OLLAMA_TOP_K)
      --provider-ollama-min-p=0                               Min P ($PROVIDER_OLLAMA_MIN_P)
      --provider-ollama-no-think                              Disable thinking ($PROVIDER_OLLAMA_NO_THINK)
      --provider-gemini-model="gemini-2.0-flash"              Gemini model name ($PROVIDER_GEMINI_MODEL)
      --provider-gemini-token=STRING                          Google AI API key ($PROVIDER_GEMINI_TOKEN)
      --provider-gemini-timeout=120s                          Timeout ($PROVIDER_GEMINI_TIMEOUT)
      --provider-gemini-max-tokens=8192                       Max tokens ($PROVIDER_GEMINI_MAX_TOKENS)
      --provider-gemini-temperature=0.8                       Temperature ($PROVIDER_GEMINI_TEMPERATURE)
      --provider-gemini-top-p=0.9                             Top P ($PROVIDER_GEMINI_TOP_P)
      --provider-gemini-top-k=40                              Top K ($PROVIDER_GEMINI_TOP_K)
      --provider-gemini-threshold-harassment="NONE"           Harassment threshold ($PROVIDER_GEMINI_THRESHOLD_HARASSMENT)
      --provider-gemini-threshold-hate-speech="NONE"          Hate speech threshold ($PROVIDER_GEMINI_THRESHOLD_HATE_SPEECH)
      --provider-gemini-threshold-sexually-explicit="NONE"    Explicit content ($PROVIDER_GEMINI_THRESHOLD_EXPLICIT)
      --provider-gemini-threshold-dangerous-content="NONE"    Dangerous content threshold ($PROVIDER_GEMINI_THRESHOLD_DANGEROUS_CONTENT)
      --oidc-enabled                                          Enable OIDC ($OIDC_ENABLED)
      --oidc-issuer=STRING                                    Issuer URL ($OIDC_ISSUER)
      --oidc-client-id=STRING                                 Client ID ($OIDC_CLIENT_ID)
      --oidc-client-secret=STRING                             Client secret ($OIDC_CLIENT_SECRET)
      --oidc-gc=5m                                            GC interval for expired sessions ($OIDC_GC)
      --tls-enabled                                           Enable TLS ($TLS_ENABLED)
      --tls-key-file="/etc/tls/tls.key"                       Key file ($TLS_KEY)
      --tls-cert-file="/etc/tls/tls.crt"                      Certificate file ($TLS_CERT)
```

</details>

## Providers

The recommended configuration is by environment variables. For CLI flags see references above.

### Ollama

- ENV: `PROVIDER_TYPE=ollama`

**Notes:**

This is the default provider. You have to download models before you can use them. For example: `ollama pull qwen3:14b`.

**Configuration:**

| ENV                            | Default value            | Description        |
|--------------------------------|--------------------------|--------------------|
| `PROVIDER_OLLAMA_URL`          | `http://localhost:11434` | Ollama URL         |
| `PROVIDER_OLLAMA_MODEL`        | `qwen3:14b`              | Default model name |
| `PROVIDER_OLLAMA_TIMEOUT`      | `120s`                   | Timeout            |
| `PROVIDER_OLLAMA_CONTEXT_SIZE` | `32768`                  | Context size       |
| `PROVIDER_OLLAMA_MAX_TOKENS`   | `32768`                  | Max tokens         |
| `PROVIDER_OLLAMA_TEMPERATURE`  | `0.6`                    | Temperature        |
| `PROVIDER_OLLAMA_TOP_P`        | `0.95`                   | Top P              |
| `PROVIDER_OLLAMA_TOP_K`        | `20`                     | Top K              |
| `PROVIDER_OLLAMA_MIN_P`        | `0`                      | Min P              |
| `PROVIDER_OLLAMA_NO_THINK`     | `false`                  | Disable thinking   |

### OpenAI

- ENV: `PROVIDER_TYPE=openai`

**Notes:**

Any OpenAI-compatible provider can be used. This includes services like OpenRouter, DeepInfra, LiteLLM, etc. You can
specify the provider's endpoint by setting the `PROVIDER_OPENAI_URL` environment variable.

**Configuration:**

| ENV                           | Default value               | Description        |
|-------------------------------|-----------------------------|--------------------|
| `PROVIDER_OPENAI_URL`         | `https://api.openai.com/v1` | OpenAI base URL    |
| `PROVIDER_OPENAI_MODEL`       | `gpt-4o`                    | Default model name |
| `PROVIDER_OPENAI_TOKEN`       |                             | OpenAI API token   |
| `PROVIDER_OPENAI_TIMEOUT`     | `3m`                        | Timeout            |
| `PROVIDER_OPENAI_MAX_TOKENS`  | `8192`                      | Max tokens         |
| `PROVIDER_OPENAI_TEMPERATURE` | `0.8`                       | Temperature        |
| `PROVIDER_OPENAI_TOP_P`       | `0.9`                       | Top P              |

**Examples:**

OpenAI

    docker run --rm -v "$(pwd):/data" -p 8080:8080 \
      -e PROVIDER_TYPE=openai \
      -e PROVIDER_OPENAI_TOKEN=sk-SUPER-SECRET-TOKEN \
      ghcr.io/reddec/dreaming-bard:latest

DeepInfra

    docker run --rm -v "$(pwd):/data" -p 8080:8080 \
      -e PROVIDER_TYPE=openai \
      -e PROVIDER_OPENAI_URL=https://api.deepinfra.com/v1/openai
      -e PROVIDER_OPENAI_TOKEN=SUPER-SECRET-TOKEN \
      ghcr.io/reddec/dreaming-bard:latest

OpenRouter

    docker run --rm -v "$(pwd):/data" -p 8080:8080 \
      -e PROVIDER_TYPE=openai \
      -e PROVIDER_OPENAI_URL=https://openrouter.ai/api/v1
      -e PROVIDER_OPENAI_TOKEN=SUPER-SECRET-TOKEN \
      ghcr.io/reddec/dreaming-bard:latest

### Gemini

- ENV: `PROVIDER_TYPE=gemini`

**Notes:**

This provider uses the Google AI (Gemini) API. You will need to obtain an API key from Google AI Studio.

**Configuration:**

| ENV                                           | Default value      | Description                 |
|-----------------------------------------------|--------------------|-----------------------------|
| `PROVIDER_GEMINI_MODEL`                       | `gemini-1.5-flash` | Default model name          |
| `PROVIDER_GEMINI_TOKEN`                       |                    | Google AI API key           |
| `PROVIDER_GEMINI_TIMEOUT`                     | `120s`             | Timeout                     |
| `PROVIDER_GEMINI_MAX_TOKENS`                  | `8192`             | Max tokens                  |
| `PROVIDER_GEMINI_TEMPERATURE`                 | `0.8`              | Temperature                 |
| `PROVIDER_GEMINI_TOP_P`                       | `0.9`              | Top P                       |
| `PROVIDER_GEMINI_TOP_K`                       | `40`               | Top K                       |
| `PROVIDER_GEMINI_THRESHOLD_HARASSMENT`        | `NONE`             | Harassment threshold        |
| `PROVIDER_GEMINI_THRESHOLD_HATE_SPEECH`       | `NONE`             | Hate speech threshold       |
| `PROVIDER_GEMINI_THRESHOLD_SEXUALLY_EXPLICIT` | `NONE`             | Explicit content threshold  |
| `PROVIDER_GEMINI_THRESHOLD_DANGEROUS_CONTENT` | `NONE`             | Dangerous content threshold |

**Example:**

    docker run --rm -v "$(pwd):/data" -p 8080:8080 \
      -e PROVIDER_TYPE=gemini \
      -e PROVIDER_GEMINI_TOKEN=A-super-secret-x \
      ghcr.io/reddec/dreaming-bard:latest

## OIDC and SSO

No SSO-[tax](http://sso.tax/). Anyway, it's single-tenant.

**Configuration:**

| ENV                  | Default value | Description                      |
|----------------------|---------------|----------------------------------|
| `OIDC_ENABLED`       | `false`       | Enable OIDC                      |
| `OIDC_ISSUER`        |               | Issuer URL                       |
| `OIDC_CLIENT_ID`     |               | Client ID                        |
| `OIDC_CLIENT_SECRET` |               | Client secret                    |
| `OIDC_GC`            | `5m`          | GC interval for expired sessions |

**Example**

    docker run --rm -v "$(pwd):/data" -p 8080:8080 \
      -e OIDC_ENABLED=true \
      -e OIDC_ISSUER=https://zitadel.example.com \
      -e OIDC_CLIENT_ID=my-client-id \
      -e OIDC_CLIENT_SECRET=aaaaafooobaaar \
      ghcr.io/reddec/dreaming-bard:latest

## TLS

Yes, it has. No, there is no integration with Let's encrypt & co.

**Configuration:**

| ENV             | Default value      | Description      |
|-----------------|--------------------|------------------|
| `TLS_ENABLED`   | `false`            | Enable TLS       |
| `TLS_KEY_FILE`  | `/etc/tls/tls.key` | Key file         |
| `TLS_CERT_FILE` | `/etc/tls/tls.crt` | Certificate file |

## License

GPLv3 - See [LICENSE](LICENSE) for full terms.

**TL;DR**: Use Dreaming-Bard freely as a service.
If you modify it or include its code in your project, you must open-source those parts under GPLv3.
