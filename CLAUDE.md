# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Orion is a real-time multilingual translation system for live events, enabling speakers to broadcast their speech with real-time translation to multiple languages for viewers. The system consists of:

- **Three Vue3 frontends**: Speaker web app, Viewer web app (mobile-first), Admin web app
- **Go backend**: REST API + WebSocket server integrating Google Speech-to-Text and Translation APIs
- **External services**: Google Cloud STT (Streaming mode) and Translation API
- **Deployment**: AWS EC2 with Docker containers orchestrated via docker-compose

## Development Commands

### Frontend (Vue3 + Vite + TypeScript)
```bash
# Development for each frontend app
cd front-end/apps/speaker-web && pnpm dev
cd front-end/apps/viewer-web && pnpm dev
cd front-end/apps/admin-web && pnpm dev

# Build for production
pnpm build

# Run tests
pnpm test

# Lint and format
pnpm lint
pnpm format
```

### Backend (Go + Gin)
```bash
# Run locally
go run cmd/server/main.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build binary
go build -o bin/orion-server cmd/server/main.go

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f backend
```

## Architecture

### System Flow
```
Speaker Browser ──┐
                  │ WebSocket Audio Stream
Viewer Browser ───┼─────> Go Backend ──> Google STT
                  │           │
Admin Browser ────┘           ├──> Google Translation
                              │
                              └──> Cache (Subtitles, Tokens, QR Status)
```

### Backend Architecture (Layered)

- **Interface Layer** (接口层): HTTP routing, WebSocket protocol, authentication middleware
- **Application Layer** (应用层): Activity management, translation pipeline, subtitle broadcasting, QR code management
- **Domain Layer** (领域层): Core entities (Activity, Sentence, Subscription, Token, ViewerEntry)
- **Infrastructure Layer** (基础设施层): Google API adapters, cache (Redis/in-memory), file storage, QR code generator, configuration

### Backend Modules

1. **Activity Management**: CRUD operations, status transitions (draft/published/closed), QR code generation
2. **Authentication**: Admin JWT login, speaker/viewer token generation and validation
3. **Real-time Translation**:
   - `AudioIngestor`: Receives audio chunks via WebSocket
   - `STTClient`: Google Streaming Speech-to-Text integration
   - `TranslationClient`: Google Translation API for multi-language output
   - `SubtitleDispatcher`: Broadcasts subtitles to viewers by language subscription
   - History cache: Last 5 minutes of subtitles (Redis or in-memory ring buffer)
4. **File & Resource**: Cover image uploads (local filesystem initially, S3 later)
5. **QR Code**: Generate viewer entry QR codes, revoke/activate based on activity status

### Frontend Architecture (Monorepo with pnpm workspace)

```
front-end/
  packages/
    shared-ui/       # Reusable Vue components
    shared-utils/    # WebSocket SDK, language configs, utilities
  apps/
    speaker-web/     # Speaker interface
    viewer-web/      # Viewer interface (mobile-first)
    admin-web/       # Admin dashboard
```

**Key Frontend Patterns**:
- Composition API with `<script setup>` syntax
- Pinia for state management
- Unified WebSocket SDK in `shared-utils` with auto-reconnect, heartbeat, event subscription
- Mobile-first responsive design for viewer app (designed for QR code scanning)

### WebSocket Protocol

**Speaker → Backend**:
- `AUTH`: Authentication with activity ID and input language
- `AUDIO`: Binary audio chunks (PCM/Opus, 16kHz)
- `CONTROL`: Start/stop commands

**Backend → Viewer**:
- `SUBTITLE`: Real-time translated text with sentence ID, language, text, timestamp
- `HISTORY`: Recent subtitles (last 5 minutes) on connection
- `STATE`: Connection status, errors, reconnect guidance
- `PING/PONG`: Heartbeat every 30 seconds

### API Design

- **Base URL**: `/api/v1`
- **Authentication**: JWT in `Authorization: Bearer <token>` header
- **Key endpoints**:
  - `POST /auth/login` - Admin login
  - `GET /activities` - List activities
  - `POST /activities` - Create activity (returns viewer entry QR code)
  - `POST /activities/{id}/publish` - Publish activity
  - `POST /activities/{id}/close` - Close activity (auto-revokes QR code)
  - `GET /activities/{id}/viewer-entry` - Get QR code (PNG/SVG/Base64)
  - `POST /activities/{id}/viewer-entry/revoke` - Manually revoke QR code
  - `POST /activities/{id}/tokens/speaker` - Generate speaker token
  - `POST /activities/{id}/tokens/viewer` - Generate viewer invite code
  - `POST /uploads/cover` - Upload cover image

## Important Implementation Notes

### Audio Processing
- Frontend captures microphone audio via `MediaDevices.getUserMedia`
- Use Web Audio API + Web Worker for sampling and encoding
- Audio chunks: 100-200ms packets (PCM 16kHz or Opus compressed)
- Backend streams directly to Google STT Streaming API

### Translation Pipeline
1. Speaker sends audio via WebSocket
2. Backend creates Google STT streaming session
3. STT returns final recognized text (with `isFinal` flag)
4. Backend calls Google Translation API for target languages
5. Backend broadcasts translated subtitles to viewers by language subscription
6. Subtitles cached for 5-minute history window

### QR Code Workflow
- **Generation**: Automatic on activity creation, displayed on success page
- **Format**: `{VIEWER_BASE_URL}/activity/{activityId}?code={inviteCode}`
- **Library**: Use `github.com/skip2/go-qrcode` for PNG/SVG generation
- **Revocation**: Automatic on activity close, manual via admin
- **Display**: Backend returns Base64-encoded image for frontend rendering

### Mobile-First Viewer Design
- Viewer app optimized for QR code scanning on mobile browsers
- Virtual list for efficient subtitle rendering
- Auto-scroll on new subtitles (with manual override)
- Font size adjustment and dark mode (future enhancement)
- Keep-screen-awake handling

## Configuration

### Backend `.env`
```bash
APP_PORT=8080
APP_ENV=production
JWT_SECRET_PATH=/secrets/jwt_private.pem
GOOGLE_APPLICATION_CREDENTIALS=/secrets/google-service-account.json
REDIS_URL=redis://redis:6379/0
WS_PING_INTERVAL=30s
HISTORY_CACHE_TTL=5m
VIEWER_BASE_URL=https://orion.example.com
```

### Frontend `.env.[mode]`
```bash
VITE_API_BASE_URL=https://api.orion.example.com
VITE_WS_URL=wss://api.orion.example.com/ws
VITE_VIEWER_BASE_URL=https://orion.example.com
```

## Deployment

- **Platform**: AWS EC2 Ubuntu 22.04 LTS (t3.medium minimum)
- **Containers**: Backend + Redis via docker-compose
- **Reverse Proxy**: Nginx for TLS termination, static files, WebSocket upgrade
- **HTTPS**: Let's Encrypt via Certbot with auto-renewal
- **Frontend**: Static files served by Nginx (can migrate to S3 + CloudFront later)

### Deployment Steps
1. Prepare EC2 instance with security groups (ports 80, 443, 22)
2. Install Docker, docker-compose, Nginx, Git
3. Clone repository to `/opt/orion`
4. Configure secrets in `/opt/orion/secrets` (Google service account JSON, JWT keys)
5. Build frontends and copy `dist` to `/var/www` directories
6. Build backend Docker image: `docker-compose build`
7. Start services: `docker-compose up -d`
8. Configure Nginx with provided config template
9. Apply HTTPS certificate: `certbot --nginx -d orion.example.com`
10. Verify: speaker audio streaming, admin QR code generation/download, viewer mobile display

## Testing Strategy

- **Unit Tests**: Frontend (Vitest + Vue Testing Library), Backend (testify)
- **Integration Tests**: Cypress/Playwright for end-to-end flows including QR code generation/revocation
- **Performance Tests**: k6/Locust for WebSocket concurrency and translation latency
- **Target**: Translation completion ≤5s, error rate ≤1%, core coverage ≥70%
- **Mobile Testing**: Browser DevTools + limited real device testing (full mobile coverage in future iterations)

## Security Requirements

- All communication over HTTPS/WSS
- JWT authentication with RS256 asymmetric signing
- Speaker/viewer tokens: short-lived JWT (1-2 hours)
- Admin tokens: 2-hour access + 7-day refresh token
- Google API credentials: service account JSON stored in secure directory or AWS Secrets Manager
- Rate limiting on sensitive endpoints (token generation, uploads, QR operations)
- QR code links tied to activity status (revoked when activity closes)
- No Google API keys exposed to frontend

## Google Cloud Integration

### Speech-to-Text
- Use `StreamingRecognize` with language code from speaker
- Enable automatic punctuation for complete sentences
- Only process `isFinal=true` results
- Audio specs: LINEAR16 or OGG_OPUS, 16000 Hz

### Translation API
- Call `TranslateText` for target languages configured in activity
- Cache language support list in backend
- Retry 3 times on quota exceeded or network errors
- Provide user-friendly error messages on frontend

### Credentials
- Service account JSON file
- Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable
- Store credentials in secure EC2 directory or Secrets Manager

## Error Handling

### Backend Error Codes
- `UNAUTHORIZED` (401): Authentication failed
- `FORBIDDEN` (403): Insufficient permissions
- `ACTIVITY_NOT_FOUND` (404): Activity doesn't exist
- `ACTIVITY_CLOSED` (409): Activity has ended
- `INVALID_LANGUAGE` (400): Unsupported target language
- `GOOGLE_STT_ERROR` (502): Speech-to-Text API failure
- `GOOGLE_TRANSLATE_ERROR` (502): Translation API failure
- `QR_GENERATE_FAILED` (500): QR code generation error
- `RATE_LIMITED` (429): API rate limit exceeded

### Response Format
```json
{
  "code": "ERROR_CODE",
  "message": "Error description",
  "data": null
}
```

## Monitoring & Logging

- **Backend Logging**: Zap structured JSON logs (request ID, activity ID, user role, errors, QR operations)
- **Key Metrics**:
  - WebSocket connection count
  - Translation success rate and latency
  - STT/Translation API response times
  - QR code generation/download count
  - Viewer entry failure rate
- **Future**: Prometheus + Grafana or AWS CloudWatch integration

## Code Style

- **All documentation, comments, and UI text in Chinese** (with i18n support for future expansion)
- **Frontend**: ESLint + Prettier + Stylelint for unified style
- **Backend**: Follow Go standard conventions (`gofmt`, `golint`)
- **Naming**: Use descriptive English names for code, Chinese for user-facing content

## Future Scalability

- Horizontal scaling: Multiple backend instances + Redis Pub/Sub for subtitle broadcasting
- Database persistence: PostgreSQL/Firestore for activities, users, subtitles
- Message queue: NATS/Kafka for complex translation pipelines
- Advanced features: Short URLs, SMS notifications, batch QR code export
