# Ollama pull client

Pulls model from the Ollama repository without the Ollama server instance running.

In my use case, I run it once in a kubernetes job to prepare a shared pvc.

## Usage
With an existing ssh id key at the default position (`~/.ollama`)
```shell
opull  all-minilm 
```
Note: [`all-minilm`](https://ollama.com/library/all-minilm) is a small (45mb) model to test the download
