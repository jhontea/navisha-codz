# API Versioning Strategy

## Overview

The Coding Challenge Platform API uses **URL-based versioning** with backward compatibility support. This document outlines the versioning strategy, migration paths, and deprecation policies.

## Versioning Scheme

### URL Path Versioning
All API endpoints are versioned via the URL path prefix:

```
/api/v1/<resource>
/api/v2/<resource>  (future)
```

### Current Versions

| Version | Status     | Release Date | Sunset Date |
|---------|------------|-------------|-------------|
| v1      | Active     | 2025-01-01  | TBD         |
| v2      | Planned    | 2026-Q1     | TBD         |

### Accessing APIs

```bash
# Versioned access (recommended)
curl https://api.coding-challenge.com/api/v1/problems

# Unversioned access (backward compatible, deprecated)
curl https://api.coding-challenge.com/api/problems
```

## Version Headers

### Request Headers

| Header               | Description                                    | Example               |
|----------------------|------------------------------------------------|-----------------------|
| `Accept-Version`     | Preferred API version (overrides URL version)  | `v1`, `v2`           |
| `X-API-Version`      | Explicit API version request                   | `1`, `2`             |

### Response Headers

| Header             | Description                                      | Example                |
|--------------------|--------------------------------------------------|------------------------|
| `X-API-Version`    | API version used to process the request          | `v1`, `v2`            |
| `X-API-Deprecated` | Indicates the version is deprecated (`true`)     | `true`, `false`       |
| `Sunset`           | Date when the version will be removed (RFC 1123) | `Sat, 31 Dec 2026 23:59:59 GMT` |
| `Deprecation`      | Deprecation status                               | `true`                |

## Backward Compatibility

### Unversioned Endpoints
Legacy `/api/<resource>` endpoints are maintained for backward compatibility:
- Automatically route to the latest stable version (currently v1)
- Return `X-API-Deprecated: true` header
- Will be removed after a 12-month deprecation period

### Request Forwarding
The API Gateway automatically strips the version prefix before forwarding to backend services:
- `/api/v1/problems` → `/problems` (to backend)
- `/api/v2/problems` → `/problems` (to backend)

## Deprecation Policy

### Sunset Header
When a version is deprecated, the API returns:
```
Sunset: Sat, 31 Dec 2026 23:59:59 GMT
Deprecation: true
```

### Deprecation Timeline

```
Announcement → 6 months → Deprecation (warnings) → 6 months → Sunset (removal)
```

| Phase           | Duration | Behavior                                                    |
|-----------------|----------|-------------------------------------------------------------|
| Active          | -        | Full support                                                |
| Deprecated      | 6 months | Returns `Deprecation: true` + `Sunset` headers, still works |
| Sunset          | -        | Returns HTTP 410 Gone                                       |

### Migration Guide

#### From Unversioned to v1
1. Change API path from `/api/problems` to `/api/v1/problems`
2. Update any hardcoded URL references
3. Verify response headers include `X-API-Version: v1`

#### From v1 to v2 (Future)
1. Update all API paths from `/api/v1/` to `/api/v2/`
2. Review changelog for breaking changes
3. Update request/response handling for new formats
4. Test in staging before production deployment

## Breaking Changes Policy

### What Constitutes a Breaking Change
- Removal or renaming of fields in response JSON
- Changes to request parameter requirements
- Changes to authentication/authorization requirements
- Changes to error response format
- Removal of endpoints

### Non-Breaking Changes
- Adding new fields to responses
- Adding new endpoints
- Adding optional request parameters
- Performance improvements
- Bug fixes that don't change contract

## Version Management

### Adding a New Version (v2)
1. Create new route group in gateway:
```go
v2 := api.Group("/v2")
// Register v2 routes
```

2. Update service routes mapping in the gateway
3. Add version info to `/version` endpoint
4. Announce deprecation of v1
5. Maintain both versions until v1 is sunset

### Version Discovery
```bash
# List available versions
GET /version

Response:
{
  "versions": ["v1", "v2"],
  "current": "v1",
  "deprecated": [],
  "sunset": null
}
```

## Gateway Implementation

The API Gateway handles version routing automatically:

```go
// Version routing in services/api-gateway-golang/main.go
gw.serviceRoutes["/api/v1/problems"] = "problem-service"
gw.serviceRoutes["/api/v2/problems"] = "problem-service" // future

// Version response header
c.Header("X-API-Version", "v1")
```

## Testing Versioning

```bash
# Test v1 endpoint
curl -H "Authorization: Bearer <token>" https://api.coding-challenge.com/api/v1/problems

# Test unversioned (backward compatible)
curl -H "Authorization: Bearer <token>" https://api.coding-challenge.com/api/problems

# Check version headers
curl -I -H "Authorization: Bearer <token>" https://api.coding-challenge.com/api/v1/problems
```