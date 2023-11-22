# money-tracker-telegram

![Version: 0.3.0](https://img.shields.io/badge/Version-0.3.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.3.0](https://img.shields.io/badge/AppVersion-0.3.0-informational?style=flat-square)

A Helm chart for Kubernetes to deploy the money tracker telegram bot

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | postgresql | 13.2.15 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.allowedUser | string | `"<your telegram username>"` | Telegram username handle, must be filled in. |
| config.apiKey | string | `"<telegram bot API key here>"` | Telegram API key, must be filled in. |
| config.debug | bool | `false` | Enable debug logging |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"Always"` | Image pull policy in Kubernetes |
| image.repository | string | `"ghcr.io/ashwinath/money-tracker-telegram"` | Respository of the image. |
| image.tag | string | `"latest"` | Override this value for the desired image tag |
| nameOverride | string | `""` |  |
| podAnnotations | object | `{}` |  |
| postgresql.auth.postgresPassword | string | `"password"` | Password for postgresql database, highly recommended to change this value |
| postgresql.primary.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim  |
| replicas | int | `1` |  |
| resources | object | `{}` | Resources requests and limits for the application |
| service.port | int | `80` |  |
